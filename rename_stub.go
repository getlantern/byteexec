// +build !windows,!darwin

package byteexec

import (
	"path"
)

func renameExecutable(orig string) string {
	return orig
}

func pathForRelativeFiles() string {
	return inHomeDir(".byteexec")
}
