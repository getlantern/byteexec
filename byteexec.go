// Package byteexec provides a very basic facility for running executables
// supplied as byte arrays, which is handy when used with
// github.com/jteeuwen/go-bindata.
//
// byteexec works by storing the provided command in a file.
//
// Example Usage:
//
//    programBytes := // read bytes from somewhere
//    be, err := byteexec.New(programBytes)
//    if err != nil {
//      log.Fatalf("Uh oh: %s", err)
//    }
//    cmd := be.Command("arg1", "arg2")
//    // cmd is an os/exec.Cmd
//    err = cmd.Run()
//
// Note - byteexec.New is somewhat expensive,
package byteexec

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/getlantern/golog"
)

const (
	fileMode = 0755
)

var (
	log = golog.LoggerFor("Exec")

	initMutex sync.Mutex
)

// Exec is a handle to an executable that can be used to create an exec.Cmd
// using the Command method. Exec is safe for concurrent use.
type Exec struct {
	filename string
}

// New creates a new Exec using the program stored in the provided data, at the
// provided filename (relative or absolute path allowed). This can be a somewhat
// expensive, so it's best to create only one Exec per executable and reuse
// that.
//
// WARNING - if a file already exists at this location and its contents differ
// from data, Exec will attempt to overwrite it.
func New(data []byte, filename string) (*Exec, error) {
	// Use initMutex to synchronize file operations by this process
	initMutex.Lock()
	defer initMutex.Unlock()

	filename = renameExecutable(filename)
	log.Tracef("Renamed executable to %s for this platform", filename)

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, fileMode)
	if err != nil {
		if !os.IsExist(err) {
			return nil, fmt.Errorf("Unexpected error opening %s: %s", filename, err)
		}

		log.Tracef("%s already exists, check to make sure contents is the same", filename)
		if checksumsMatch(filename, data) {
			return newExecFromExisting(filename)
		}

		log.Tracef("Data in %s doesn't match expected, truncating file", filename)
		file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
		if err != nil {
			return nil, fmt.Errorf("Unable to truncate %s: %s", err)
		}
	}

	log.Tracef("Creating new file at %s", filename)
	_, err = file.Write(data)
	if err != nil {
		os.Remove(filename)
		return nil, fmt.Errorf("Unable to write to file at %s: %s", filename, err)
	}
	file.Sync()
	file.Close()
	return newExec(filename)
}

// Command creates an exec.Cmd using the supplied args.
func (be *Exec) Command(args ...string) *exec.Cmd {
	return exec.Command(be.filename, args...)
}

func checksumsMatch(filename string, data []byte) bool {
	shasum := sha256.New()
	file, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		log.Tracef("Unable to open existing file at %s for reading: %s", filename, err)
		return false
	}
	_, err = io.Copy(shasum, file)
	if err != nil {
		log.Tracef("Unable to read bytes to calculate sha sum: %s", err)
		return false
	}
	checksumOnDisk := shasum.Sum(nil)
	expectedChecksum := sha256.Sum256(data)
	return bytes.Equal(checksumOnDisk, expectedChecksum[:])
}

func newExecFromExisting(filename string) (*Exec, error) {
	log.Tracef("Data in %s matches expected, using existing", filename)
	fi, err := os.Stat(filename)
	if err != nil || fi.Mode() != fileMode {
		log.Tracef("Chmodding %s", filename)
		err = os.Chmod(filename, fileMode)
		if err != nil {
			return nil, fmt.Errorf("Unable to chmod file %s: %s", filename, err)
		}
	}
	return newExec(filename)
}

func newExec(filename string) (*Exec, error) {
	absolutePath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	return &Exec{filename: absolutePath}, nil
}
