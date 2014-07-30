bin [![GoDoc](https://godoc.org/github.com/rjeczalik/bin?status.svg)](https://godoc.org/github.com/rjeczalik/bin) [![Build Status](https://travis-ci.org/rjeczalik/bin.png?branch=master)](https://travis-ci.org/rjeczalik/bin "linux_amd64") [![Build Status](https://travis-ci.org/rjeczalik/bin.png?branch=osx)](https://travis-ci.org/rjeczalik/bin "darwin_amd64") [![Build status](https://ci.appveyor.com/api/projects/status/sl6pjb76vk3uw4s2)](https://ci.appveyor.com/project/rjeczalik/bin "windows_amd64")
=========

Searches for Go executables in $PATH / $GOBIN / $GOPATH and updates them.

**NOTE** Go version 1.3 required.

## cmd/gobin [![GoDoc](https://godoc.org/github.com/rjeczalik/bin/cmd/gobin?status.png)](https://godoc.org/github.com/rjeczalik/bin/cmd/gobin)

*Installation*

```
~ $ go get -u github.com/rjeczalik/bin/cmd/gobin
```

*Documentation*

[godoc.org/github.com/rjeczalik/bin/cmd/gobin](http://godoc.org/github.com/rjeczalik/bin/cmd/gobin)

*Usage*

```bash
~ $ GOPATH=~ go get github.com/rjeczalik/pkgconfig/cmd/pkg-config \
                    github.com/rjeczalik/bindata/cmd/bindata      \
                    github.com/rjeczalik/which/cmd/gowhich        \
                    github.com/rjeczalik/tools/cmd/gotree
```
```bash
~ $ GOPATH=~ gobin
/home/rjeczalik/bin/bindata	(github.com/rjeczalik/bindata/cmd/bindata)
/home/rjeczalik/bin/gotree	(github.com/rjeczalik/tools/cmd/gotree)
/home/rjeczalik/bin/gowhich	(github.com/rjeczalik/which/cmd/gowhich)
/home/rjeczalik/bin/pkg-config	(github.com/rjeczalik/pkgconfig/cmd/pkg-config)
```
```bash
~ $ GOPATH=~ gobin -u
ok	/home/rjeczalik/bin/gowhich	(github.com/rjeczalik/which/cmd/gowhich)	5.926s
ok	/home/rjeczalik/bin/gotree	(github.com/rjeczalik/tools/cmd/gotree)	2.494s
ok	/home/rjeczalik/bin/pkg-config	(github.com/rjeczalik/pkgconfig/cmd/pkg-config)	2.635s
ok	/home/rjeczalik/bin/bindata	(github.com/rjeczalik/bindata/cmd/bindata)	3.474s
```

**NOTE** Bumping GOMAXPROCS value up may speed up gobin significantly.
