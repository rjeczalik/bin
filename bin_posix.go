// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package bin

import (
	"os"
	"os/user"
	"strconv"
	"syscall"
)

var uid, gid uint32

func init() {
	if u, err := user.Current(); err == nil {
		if i, err := strconv.ParseUint(u.Uid, 10, 32); err == nil {
			uid = uint32(i)
		}
		if i, err := strconv.ParseUint(u.Gid, 10, 32); err == nil {
			gid = uint32(i)
		}
	}
}

const (
	s_IXUSR os.FileMode = 0x40
	s_IXGRP os.FileMode = 0x08
	s_IXOTH os.FileMode = 0x01
)

func isExecutable(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	mode := fi.Mode()
	if (mode & s_IXOTH) != 0 {
		return true
	}
	sys := fi.Sys().(*syscall.Stat_t)
	return (sys.Gid == gid && (mode&s_IXGRP) != 0) ||
		(sys.Uid == uid && (mode&s_IXUSR) != 0)
}
