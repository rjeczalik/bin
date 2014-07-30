// cmd/gobin TODO(rjeczalik): document
//
// Usage
//
//   NAME:
//       gobin - searches for Go executables in $PATH/$GOBIN/$GOPATH
//               and lists or updates them.
//
//   USAGE:
//       gobin [-u] [-v] [path|package...]
//
//   FLAGS:
//       -v  Turns on verbose output
//       -u  Updates Go binaries
//
//   EXAMPLES:
//       gobin                    Lists all Go binaries (looks up $PATH/$GOBIN/$GOPATH)
//       gobin -v -u              Updates all Go binaries
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
	gobin [-u] [-v] [path|package...]

FLAGS:
	-v  Turns on verbose output
	-u  Updates Go binaries

EXAMPLES:
	gobin                    Lists all Go binaries (looks up $PATH/$GOBIN/$GOPATH)
	gobin -v -u              Updates all Go binaries
	gobin -u github.com      Updates all Go binaries installed from github.com
	gobin ~/bin              Lists all Go binaries from the ~/bin directory`

var (
	update  bool
	verbose bool
)

func ishelp(s string) bool {
	return s == "-h" || s == "-help" || s == "help" || s == "--help" || s == "/?"
}

func parse() []string {
	flag.Usage = func() { die(usage) }
	flag.BoolVar(&update, "u", false, "")
	flag.BoolVar(&verbose, "v", false, "")
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

// TODO(rjeczalik): Bin.CanWrite needs a Flock here
func main() {
	if len(os.Args) == 2 && ishelp(os.Args[1]) {
		fmt.Println(usage)
		return
	}
	var (
		b []bin.Bin
		s map[string][]bin.Bin
		e error
		a = parse()
	)
	if verbose {
		b, s, e = bin.SearchSymlink(a)
	} else {
		b, e = bin.Search(a)
	}
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
		bin.Update(b)
	} else {
		for i := range b {
			fmt.Printf("%s\t(%s)\n", b[i].Path, b[i].Package)
		}
	}
	_ = s
}
