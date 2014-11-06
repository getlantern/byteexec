// Package byteexec provides a very basic facility for running executables
// supplied as byte arrays, which is handy when used with
// github.com/jteeuwen/go-bindata.
//
// ByteExec works by storing the provided command in a file.
//
// Example Usage:
//
//    programBytes := // read bytes from somewhere
//    be, err := NewByteExec(programBytes)
//    if err != nil {
//      log.Fatalf("Uh oh: %s", err)
//    }
//    defer be.Close()
//    cmd := be.Command("arg1", "arg2")
//    // cmd is an os/exec.Cmd
//    err = cmd.Run()
package byteexec

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	log = golog.LoggerFor("byteexec")

	initMutex sync.Mutex
)

type ByteExec struct {
	filename string
}

// New creates a new ByteExec using the program stored in the provided data, at
// the provided filename (relative or absolute path allowed).
//
// WARNING - if a file already exists at this location and its contents differ
// from data, byteexec will attempt to overwrite it.
func New(data []byte, filename string) (*ByteExec, error) {
	// Use initMutex to synchronize file operations by this process
	initMutex.Lock()
	defer initMutex.Unlock()

	filename = renameExecutable(filename)
	log.Tracef("Renamed executable to %s for this platform", filename)

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, fileMode)
	if err == nil {
		log.Tracef("Creating new file at %s", filename)
	} else {
		if !os.IsExist(err) {
			return nil, fmt.Errorf("Unexpected error opening %s: %s", filename, err)
		}

		log.Tracef("%s already exists, check to make sure contents is the same", filename)
		dataOnDisk, err := ioutil.ReadFile(filename)
		if err == nil && bytes.Equal(dataOnDisk, data) {
			log.Tracef("Data in %s matches expected, using existing", filename)
			fi, err := os.Stat(filename)
			if err != nil || fi.Mode() != fileMode {
				log.Tracef("Chmodding %s", filename)
				err = os.Chmod(filename, fileMode)
				if err != nil {
					return nil, fmt.Errorf("Unable to chmod file %s: %s", filename, err)
				}
			}
			return newByteExec(filename)
		}
		log.Tracef("Data in %s doesn't match expected, truncating file", filename)
		file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
		if err != nil {
			return nil, fmt.Errorf("Unable to truncate %s: %s", err)
		}
	}

	_, err = file.Write(data)
	if err != nil {
		os.Remove(filename)
		return nil, fmt.Errorf("Unable to write to file at %s: %s", filename, err)
	}
	file.Sync()
	file.Close()
	return newByteExec(filename)
}

// Command creates an exec.Cmd using the supplied args.
func (be *ByteExec) Command(args ...string) *exec.Cmd {
	return exec.Command(be.filename, args...)
}

func newByteExec(filename string) (*ByteExec, error) {
	absolutePath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	return &ByteExec{filename: absolutePath}, nil
}
