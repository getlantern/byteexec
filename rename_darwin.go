// +build !windows

package byteexec

import (
	"fmt"
	"os"
	"path/filepath"
)

func renameExecutable(orig string) string {
	return orig
}

func pathForRelativeFiles() (string, error) {
	// Use os.UserConfigDir to get to the "Application Support" directory. This is important as we
	// may be running in a sandbox, in which case we need the "Application Support" directory in the
	// container sandbox. os.UserConfigDir will return the correct result.
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to obtain user config directory: %w", err)
	}
	return filepath.Join(cfgDir, "byteexec"), nil
}
