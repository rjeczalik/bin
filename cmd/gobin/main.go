// Command gobin updates Go commands by reading their import paths, go-getting
// their repositories and building new executables.
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
// Example output
//
//   ~ $ GOMAXPROCS=2 gobin -u
//   ok 	/home/rjeczalik/bin/gowhich	(github.com/rjeczalik/which/cmd/gowhich)	3.470s
//   ok 	/home/rjeczalik/bin/circuit	(github.com/gocircuit/circuit/cmd/circuit)	4.980s
//   ok 	/home/rjeczalik/bin/gobin	(github.com/rjeczalik/bin/cmd/gobin)	6.108s
//   ok 	/home/rjeczalik/workspace/bin/gobin	(github.com/rjeczalik/bin/cmd/gobin)	6.108s
//   ok 	/home/rjeczalik/bin/gotree	(github.com/rjeczalik/tools/cmd/gotree)	2.458s
//   ok 	/home/rjeczalik/bin/ciexec	(github.com/rjeczalik/ciexec/cmd/ciexec)	10.918s
//   ok 	/home/rjeczalik/bin/pkg-config	(github.com/rjeczalik/pkgconfig/cmd/pkg-config)	2.518s
//   ok 	/home/rjeczalik/bin/bindata	(github.com/rjeczalik/bindata/cmd/bindata)	3.584s
//   ok 	/home/rjeczalik/bin/fakerpc	(github.com/rjeczalik/fakerpc/cmd/fakerpc)	5.498s
//   ok 	/home/rjeczalik/workspace/bin/fakerpc	(github.com/rjeczalik/fakerpc/cmd/fakerpc)	5.499s
//   ok 	/home/rjeczalik/bin/golint	(github.com/golang/lint/golint)	2.973s
//   ok 	/home/rjeczalik/bin/ungocheck	(github.com/rjeczalik/ungocheck/cmd/ungocheck)	2.308s
//   ok 	/home/rjeczalik/bin/goimports	(code.google.com/p/go.tools/cmd/goimports)	13.966s
//   ok 	/home/rjeczalik/bin/godoc	(code.google.com/p/go.tools/cmd/godoc)	17.960s
//   ok 	/home/rjeczalik/bin/pulsecli	(github.com/x-formation/pulsekit/cmd/pulsecli)	13.052s
//   ok 	/home/rjeczalik/workspace/bin/pulsecli	(github.com/x-formation/pulsekit/cmd/pulsecli)	13.052s
//
// Usage
//
//   NAME:
//       gobin - searches for Go executables in $PATH/$GOBIN/$GOPATH
//               and lists or updates them.
//
//   USAGE:
//       gobin [-u] [path|package...]
//
//   FLAGS:
//       -u  Updates Go binaries
//
//   EXAMPLES:
//       gobin                    Lists all Go binaries (looks up $PATH/$GOBIN/$GOPATH)
//       gobin -u                 Updates all Go binaries
//       gobin -u github.com      Updates all Go binaries installed from github.com
//       gobin ~/bin              Lists all Go binaries from the ~/bin directory
//
//   DANGEROUS EXAMPLES:
//       gobin -u github.com/rjeczalik  Updates all Go binaries installed from github.com/rjeczalik
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
	gobin - searches for Go executables in $PATH/$GOBIN/$GOPATH
	        and lists or updates them.

USAGE:
	gobin [-u] [path|package...]

FLAGS:
	-u    Updates Go binaries

EXAMPLES:
	gobin                    Lists all Go binaries (looks up $PATH/$GOBIN/$GOPATH)
	gobin -u                 Updates all Go binaries
	gobin -u github.com      Updates all Go binaries installed from github.com
	gobin ~/bin              Lists all Go binaries from the ~/bin directory`

var update bool

func ishelp(s string) bool {
	return s == "-h" || s == "-help" || s == "help" || s == "--help" || s == "/?"
}

func parse() []string {
	flag.Usage = func() { die(usage) }
	flag.BoolVar(&update, "u", false, "")
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
	if update {
		if self := self(); self != "" {
			for i := range b {
				if b[i].Path == self {
					b[i], b = b[len(b)-1], b[:len(b)-1]
					break
				}
			}
		}
		bin.Update(b, log)
	} else {
		for i := range b {
			fmt.Printf("%s\t(%s)\n", b[i].Path, b[i].Package)
		}
	}
}
