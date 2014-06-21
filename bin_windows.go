// +build windows

package bin

import "strings"

func isExecutable(path string) bool {
	if len(path) < 5 {
		return false
	}
	ext := path[len(path)-4:]
	return strings.ToLower(ext) == ".exe"
}
