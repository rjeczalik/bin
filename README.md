bin [![GoDoc](https://godoc.org/github.com/rjeczalik/bin?status.svg)](https://godoc.org/github.com/rjeczalik/bin) [![Build Status](https://travis-ci.org/rjeczalik/bin.png?branch=master)](https://travis-ci.org/rjeczalik/bin "linux_amd64") [![Build Status](https://travis-ci.org/rjeczalik/bin.png?branch=osx)](https://travis-ci.org/rjeczalik/bin "darwin_amd64") [![Build status](https://ci.appveyor.com/api/projects/status/sl6pjb76vk3uw4s2)](https://ci.appveyor.com/project/rjeczalik/bin "windows_amd64")
=========

Package `bin` looks for Go executable system-wide (`$PATH`, `$GOBIN`, `$GOPATH`), lists them, reads their import paths, fetches their sources and updates them.

**NOTE** Go version 1.3 required.

## cmd/gobin [![GoDoc](https://godoc.org/github.com/rjeczalik/bin/cmd/gobin?status.png)](https://godoc.org/github.com/rjeczalik/bin/cmd/gobin)

*Installation*

```
~ $ go get -u github.com/rjeczalik/bin/cmd/gobin
```

*Documentation*

[godoc.org/github.com/rjeczalik/bin/cmd/gobin](http://godoc.org/github.com/rjeczalik/bin/cmd/gobin)

*Source example*

![gobin -s](https://i.imgur.com/2qs25Cg.gif "gobin -s")

*Update example*

![gobin -u](https://i.imgur.com/AEimmsY.gif "gobin -u")
