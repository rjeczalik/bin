// Command gobin  Go commands by reading their import paths, go-getting
// their repositories and building new executables.
//
// Command gobin can search for and list Go executables, create a $GOPATH workspace
// out of a binary. It can also update an executable which does not have its
// sources fetched in current $GOPATH.
//
// Source
//
// Command gobin can guess an origin of the sources used to build Go executables
// it finds on a system. It can be used to create precise mirror of the sources
// for system Go executables, without any unneeded packages.
//
//   ~ $ gobin -s /tmp/gopath godoc golint goimports gotree gowhich
//   code.google.com/p/go.tools (download)
//   github.com/rjeczalik/tools (download)
//   github.com/rjeczalik/which (download)
//   (...)
//   github.com/rjeczalik/which/cmd/gowhich
//   code.google.com/p/go.tools/cmd/godoc
//   ~ $ tree -L 3 /tmp/gopath/src/
//   /tmp/gopath/src/
//   ├── code.google.com
//   │   └── p
//   │       └── go.tools
//   └── github.com
//       └── rjeczalik
//           ├── tools
//           └── which
//
//   7 directories, 0 files
//
// Update
//
// Command gobin can update single executable or automagically discover all of
// them in directories specified in $GOPATH, $GOBIN and $GOPATH environment
// variables.
//
// Executing gobin without any arguments makes it list all Go executables found
// in $GOPATH, $GOBIN and $GOPATH.
//
// Updating multiple executables is performed on multiple goroutines, bumping
// up $GOMAXPROCS environment variable may speed up the overall run-time
// significantly.
//
// Example
//
//   ~ $ GOMAXPROCS=2 gobin -u
//   ok 	/home/rjeczalik/bin/goimports	(code.google.com/p/go.tools/cmd/goimports)	13.966s
//   ok 	/home/rjeczalik/bin/godoc	(code.google.com/p/go.tools/cmd/godoc)	17.960s
//   ok 	/home/rjeczalik/bin/pulsecli	(github.com/x-formation/pulsekit/cmd/pulsecli)	13.052s
//   ok 	/home/rjeczalik/workspace/bin/pulsecli	(github.com/x-formation/pulsekit/cmd/pulsecli)	13.052s
//
// Usage
//
//   NAME:
//       gobin - looks for Go executables system-wide ($PATH/$GOBIN/$GOPATH),
//               lists them, fetches their sources and updates them
//
//   USAGE:
//       gobin [-u] [-s=.|gopath] [path|package...]
//
//   FLAGS:
//       -u                Updates Go binaries
//       -s <dir>          Go-gets sources for Go specified binaries into <dir> $GOPATH
//                         (use '.' for current $GOPATH)
//       -ldflags=<flags>  passes "-ldflags=flags" to "go install"
//
//   EXAMPLES:
//       gobin                    Lists all Go binaries (looks up $PATH/$GOBIN/$GOPATH)
//       gobin -s=. ~/bin         Go-gets sources used to build all Go binaries in ~/bin
//                                into current $GOPATH
//       gobin -s=/var/mirror     Go-gets all sources used to build all Go binaries found
//                                on system into new /var/mirror $GOPATH
//       gobin -u                 Updates all Go binaries
//       gobin -u github.com      Updates all Go binaries installed from github.com
//       gobin ~/bin              Lists all Go binaries from the ~/bin directory
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rjeczalik/bin"
)

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

const usage = `NAME:
	gobin - performs discovery of Go executables ($PATH/$GOBIN/$GOPATH),
	        lists them, fetches their sources and updates them

USAGE:
	gobin [-u] [-s=.|gopath] [path|package...]

FLAGS:
	-u                Updates Go binaries
	-s <dir>          Go-gets sources for Go specified binaries into <dir> $GOPATH
	                  (use '.' for current $GOPATH)
	-ldflags=<flags>  passes "-ldflags=flags" to "go install"

EXAMPLES:
	gobin                    Lists all Go binaries (looks up $PATH/$GOBIN/$GOPATH)
	gobin -s=. ~/bin         Go-gets sources used to build all Go binaries in ~/bin
	                         into current $GOPATH
	gobin -s=/var/mirror     Go-gets all sources used to build all Go binaries found
	                         on system into new /var/mirror $GOPATH
	gobin -u                 Updates all Go binaries in-place
	gobin -u github.com      Updates all Go binaries installed from github.com
	gobin ~/bin              Lists all Go binaries from the ~/bin directory
	gobin -u -ldflags='-w -s'	Updates all go binaries in-place, using "go install -ldflags='-w -s'"`

var (
	source, ldflags string
	update          bool
)

func ishelp(s string) bool {
	return s == "-h" || s == "-help" || s == "help" || s == "--help" || s == "/?"
}

func parse() []string {
	flag.Usage = func() { die(usage) }
	flag.StringVar(&source, "s", "", "")
	flag.BoolVar(&update, "u", false, "")
	flag.StringVar(&ldflags, "ldflags", "", "")
	flag.Parse()
	return flag.Args()
}

func self() string {
	if strings.Contains(os.Args[0], string(os.PathSeparator)) {
		if self, err := filepath.Abs(os.Args[0]); err == nil {
			if fiself, err := os.Stat(self); err == nil {
				if fiargs, err := os.Stat(os.Args[0]); err == nil && os.SameFile(fiself, fiargs) {
					return self
				}
			}
		}
	}
	if self, err := exec.LookPath(filepath.Base(os.Args[0])); err == nil {
		return self
	}
	return ""
}

func log(b *bin.Bin, d time.Duration, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail\t%s\t(%s)\n", b.Path, b.Package)
		fmt.Fprintf(os.Stderr, "\terror: %v\n", err)
	} else {
		fmt.Printf("ok\t%s\t(%s)\t%.3fs\n", b.Path, b.Package, d.Seconds())
	}
}

// TODO(rjeczalik): Bin.CanWrite needs a Flock here
func main() {
	if len(os.Args) == 2 && ishelp(os.Args[1]) {
		fmt.Println(usage)
		return
	}
	var b, e = bin.Search(parse())
	if e != nil {
		die(e)
	}
	var installFlags []string
	switch {
	case update:
		if self := self(); self != "" {
			for i := range b {
				if b[i].Path == self {
					b[i], b = b[len(b)-1], b[:len(b)-1]
					break
				}
			}
		}
		if ldflags != "" {
			installFlags = append(installFlags, "-ldflags="+ldflags)
		}
		bin.Update(bin.UpdateOpts{Bins: b, Log: log, Flags: installFlags})
	case source != "":
		if source == "." {
			source = os.Getenv("GOPATH")
			if source == "" {
				die("bin: unable to read current $GOPATH or $GOPATH is empty")
			}
			if i := strings.Index(source, string(os.PathListSeparator)); i != -1 {
				source = source[:i]
			}
		}
		if bin.Source(b, source) != nil {
			os.Exit(1)
		}
	default:
		for i := range b {
			fmt.Printf("%s\t(%s)\n", b[i].Path, b[i].Package)
		}
	}
}
