// +build !windows,!darwin

package byteexec

import (
	"path"
)

func renameExecutable(orig string) string {
	return orig
}

func pathForRelativeFiles(homeDir string) string {
	return path.Join(homeDir, ".byteexec")
}
