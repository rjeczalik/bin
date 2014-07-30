package bin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rjeczalik/which"
)

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

var (
	home     string
	parallel = max(runtime.GOMAXPROCS(-1), runtime.NumCPU())
)

func init() {
	if u, err := user.Current(); err == nil {
		home = u.HomeDir
	}
}

func appenduniq(a *[]string) func(...string) {
	dups := make(map[string]struct{})
	return func(s ...string) {
		for _, s := range s {
			if _, ok := dups[s]; !ok {
				*a = append(*a, s)
				dups[s] = struct{}{}
			}
		}
	}
}

func splitandmap(env string, fn func(string) string) (s []string) {
	var err error
	for _, dir := range strings.Split(os.Getenv(env), string(os.PathListSeparator)) {
		if dir = fn(dir); dir != "" {
			if dir, err = filepath.Abs(dir); err == nil {
				s = append(s, dir)
			}
		}
	}
	return
}

// TODO(rjeczalik): handle directory symlinks
func searchpaths() (s []string) {
	appends := appenduniq(&s)
	if u, err := user.Current(); err == nil {
		// Use $PATH and $GOBIN directories only if they're in $HOME.
		ishome := func(dir string) string {
			if strings.HasPrefix(dir, u.HomeDir) {
				return dir
			}
			return ""
		}
		appends(splitandmap("PATH", ishome)...)
		appends(splitandmap("GOBIN", ishome)...)
	}
	isdir := func(dir string) string {
		dir = filepath.Join(dir, "bin")
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			return dir
		}
		return ""
	}
	appends(splitandmap("GOPATH", isdir)...)
	return
}

func splitdirpkg(a []string) (dirs, pkgs []string) {
	var (
		appenddirs = appenduniq(&dirs)
		appendpkgs = appenduniq(&pkgs)
	)
	for _, a := range a {
		switch {
		case a == ".":
			if wd, err := os.Getwd(); err == nil {
				appenddirs(wd)
			}
		default:
			if fi, err := os.Stat(a); err == nil {
				if !fi.IsDir() {
					a = filepath.Dir(a)
				}
				appenddirs(a)
			} else if os.IsNotExist(err) {
				appendpkgs(a)
			}
		}
	}
	return
}

// Bin TODO(rjeczalik): document
type Bin struct {
	Path     string
	Package  string
	CanWrite bool
	err      error
}

// Err TODO(rjeczalik): document
func (b Bin) Err() error { return b.err }

// BinSlice TODO(rjeczalik): document
type BinSlice []Bin

// Implements sort.Interface.
func (b BinSlice) Sort()              { sort.Sort(b) }
func (b BinSlice) Len() int           { return len(b) }
func (b BinSlice) Less(i, j int) bool { return b[i].Path < b[j].Path }
func (b BinSlice) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// CanWrite TODO(rjeczalik): document
func CanWrite(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	dir := path
	if !fi.IsDir() {
		dir = filepath.Dir(path)
	}
	f, err := ioutil.TempFile(dir, filepath.Base(path))
	if err != nil {
		return false
	}
	f.Close()
	if !fi.IsDir() {
		err, _ = os.Rename(path, f.Name()), os.Rename(f.Name(), path)
	}
	os.Remove(f.Name())
	return (err == nil)
}

// IsExecutable TODO(rjeczalik): document
func IsExecutable(path string) bool {
	return isExecutable(path)
}

// IsBinary TODO(rjeczalik): document
func IsBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	var p [32]byte
	n, err := f.Read(p[:])
	if n == 0 || (err != nil && err != io.EOF) {
		return false
	}
	// TODO(rjeczalik): detect shebang (e.g. filter out compiled Python)?
	return !strings.Contains(http.DetectContentType(p[:n]), "text/plain")
}

// Search TODO(rjeczalik): document
func Search(args []string) ([]Bin, error) {
	bin, _, err := searchSymlink(args, false)
	return bin, err
}

// SearchSymlink TODO(rjeczalik): document
func SearchSymlink(args []string) ([]Bin, map[string][]Bin, error) {
	return searchSymlink(args, true)
}

// Update TODO(rjeczalik): document
func Update(b []Bin, log func(*Bin, time.Duration, error)) {
	type kv struct {
		k string
		v []string
	}
	var (
		builds = make(map[string][]string, len(b))
		mtx    sync.Mutex // protects b
		seterr = func(t time.Time, err error, paths ...string) {
			mtx.Lock()
			for _, path := range paths {
				for i := range b {
					if b[i].Path == path {
						log(&b[i], time.Now().Sub(t), err)
						b[i].err = err
					}
				}
			}
			mtx.Unlock()
		}
		fmterr = func(err error, p []byte) error {
			return fmt.Errorf("%v\n\t%s", err, bytes.Replace(p, []byte{'\n'}, []byte("\n\t"), -1))
		}
	)
	for i := range b {
		if b[i].CanWrite {
			if v, ok := builds[b[i].Package]; ok {
				builds[b[i].Package] = append(v, b[i].Path)
			} else {
				builds[b[i].Package] = []string{b[i].Path}
			}
		}
	}
	var (
		ch = make(chan kv, len(builds))
		wg sync.WaitGroup
	)
	for k, v := range builds {
		wg.Add(1)
		ch <- kv{k: k, v: v}
	}
	for i := min(parallel, len(builds)); i > 0; i-- {
		go func() {
			for kv := range ch {
				wrk, err := ioutil.TempDir("", "gobin")
				if err != nil {
					seterr(time.Now(), err, kv.v...)
					continue
				}
				var (
					bin  = filepath.Join(wrk, "bin")
					env  = environ(wrk, bin)
					fail = func(err error, s ...string) {
						seterr(time.Now(), err, s...)
						os.RemoveAll(wrk)
						wg.Done()
					}
				)
				t := time.Now()
				build := exec.Command("go", "get", kv.k)
				build.Env = env
				if p, err := build.CombinedOutput(); err != nil {
					fail(fmterr(err, p), kv.v...)
					continue
				}
				install := exec.Command("go", "install", kv.k)
				install.Env = env
				if p, err := install.CombinedOutput(); err != nil {
					fail(fmterr(err, p), kv.v...)
					continue
				}
				exe := filepath.Join(bin, filepath.Base(kv.v[0]))
				for _, path := range kv.v {
					if err = copyfile(path, exe); err != nil {
						seterr(time.Now(), err, path)
					}
				}
				seterr(t, nil, kv.v...)
				os.RemoveAll(wrk)
				wg.Done()
			}
		}()
	}
	wg.Wait()
	close(ch)
}

func environ(gopath, gobin string) []string {
	var s = os.Environ()
	for i := range s {
		switch {
		case strings.HasPrefix(s[i], "GOPATH="):
			s[i] = "GOPATH=" + gopath
		case strings.HasPrefix(s[i], "GOBIN="):
			s[i] = "GOBIN=" + gobin
		}
	}
	return s
}

func copyfile(dst, src string) error {
	fsrc, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fsrc.Close()
	fdst, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fdst.Close()
	_, err = io.Copy(fdst, fsrc)
	return err
}

func importpkg(path string) (string, error) {
	ex, err := which.NewExec(path)
	if err != nil {
		return "", err
	}
	if ex.Type.GOOS != runtime.GOOS || ex.Type.GOARCH != runtime.GOARCH {
		return "", errors.New("bin: cross-compiling is not supported yet")
	}
	return ex.Import()
}

func searchSymlink(args []string, symlink bool) (b []Bin, s map[string][]Bin, err error) {
	type skv struct {
		k string
		v Bin
	}
	type binpath struct {
		path     string
		canwrite bool
	}
	var (
		mtx sync.Mutex // protects b
		sch chan skv
		bs  map[string]Bin
	)
	dirs, pkgs := splitdirpkg(args)
	if dirs == nil || len(dirs) == 0 {
		dirs = searchpaths()
	}
	if dirs == nil || len(dirs) == 0 {
		return nil, nil, errors.New("bin: couldn't find any search paths")
	}
	if symlink {
		s, sch, bs = make(map[string][]Bin), make(chan skv), make(map[string]Bin)
		defer close(sch)
		go func() {
			for kv := range sch {
				if v, ok := s[kv.k]; ok {
					s[kv.k] = append(v, kv.v)
				} else {
					s[kv.k] = []Bin{kv.v}
				}
			}
		}()
	}
	// TODO(rjeczalik): cap(ch) = max(count files in dirs)
	ch, wg := make(chan binpath, 128), sync.WaitGroup{}
	for i := 0; i < parallel; i++ {
		go func() {
			for p := range ch {
				pkg, err := importpkg(p.path)
				if err != nil {
					wg.Done()
					continue
				}
				mtx.Lock()
				b = append(b, Bin{Path: p.path, Package: pkg, CanWrite: p.canwrite})
				mtx.Unlock()
				wg.Done()
			}
		}()
	}
	// TODO(rjeczalik): handle file symlinks
	for _, dir := range dirs {
		fi, err := os.Open(dir)
		if err != nil {
			continue
		}
		files, err := fi.Readdir(0)
		if err != nil {
			fi.Close()
			continue
		}
		canwrite := CanWrite(dir)
		for _, fi := range files {
			path := filepath.Join(dir, fi.Name())
			if !fi.IsDir() && fi.Mode().IsRegular() && IsExecutable(path) && IsBinary(path) {
				wg.Add(1)
				ch <- binpath{path: path, canwrite: canwrite}
			}
		}
		fi.Close()
	}
	wg.Wait()
	close(ch)
	var fb = make([]Bin, 0, len(b))
	if len(pkgs) != 0 {
		for i := range b {
			for _, pkg := range pkgs {
				if strings.HasPrefix(b[i].Package, pkg) {
					fb = append(fb, b[i])
				}
			}
		}
		b, _ = fb, bs
	}
	BinSlice(b).Sort()
	return
}
