//go:build !windows

package seed

import "path/filepath"

// realPath names the directory a path leads to.
func realPath(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}
