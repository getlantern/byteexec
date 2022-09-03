// Package byteexec provides a very basic facility for running executables
// supplied as byte arrays, which is handy when used with
// github.com/jteeuwen/go-bindata.
//
// byteexec works by storing the provided command in a file.
//
// Example Usage:
//
//    programBytes := // read bytes from somewhere
//    be, err := byteexec.New(programBytes, "new/path/to/executable")
//    if err != nil {
//      log.Fatalf("Uh oh: %s", err)
//    }
//    cmd := be.Command("arg1", "arg2")
//    // cmd is an os/exec.Cmd
//    err = cmd.Run()
//
// Note - byteexec.New is somewhat expensive, and Exec is safe for concurrent
// use, so it's advisable to create only one Exec for each executable.
package byteexec

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/getlantern/filepersist"
	"github.com/getlantern/golog"
)

var (
	log = golog.LoggerFor("Exec")

	initMutex sync.Mutex
)

// NewFileMode is the mode assigned to files passed to New.
const NewFileMode os.FileMode = 0744

// Exec is a handle to an executable that can be used to create an exec.Cmd
// using the Command method. Exec is safe for concurrent use.
type Exec struct {
	Filename string
}

// New creates a new Exec using the program stored in the provided data, at the
// provided filename (relative or absolute path allowed). If the path given is
// a relative path, the executable will be placed in one of the following
// locations:
//
// On Windows - %APPDATA%/byteexec
// On OSX - ~/Library/Application Support/byteexec
// All Others - ~/.byteexec
//
// Creating a new Exec can be somewhat expensive, so it's best to create only
// one Exec per executable and reuse that.
//
// WARNING:
//	- If a file already exists at this location and its contents differ from
//    data, Exec will attempt to overwrite it.
//	- Even when the file contents match the input data, the file mode will be
//    changed to NewFileMode.
func New(data []byte, filename string) (*Exec, error) {
	log.Tracef("Creating new at %v", filename)
	return loadExecutable(filename, data)
}

// Existing is like New, but specifically for programs which already exist in
// the given file. This can be useful for situations in which it is important
// that the file not be modified (New can affect file permissions even when the
// file is not overwritten).
//
// If the path given is a relative path, the executable is assumed to be in one
// of the following locations:
//
// On Windows - %APPDATA%/byteexec
// On OSX - ~/Library/Application Support/byteexec
// All Others - ~/.byteexec
func Existing(filename string) (*Exec, error) {
	log.Tracef("Loading existing at %v", filename)
	return loadExecutable(filename, nil)
}

// If data is nil, we assume the file is to be loaded and not modified.
func loadExecutable(filename string, data []byte) (*Exec, error) {
	// Use initMutex to synchronize file operations by this process
	initMutex.Lock()
	defer initMutex.Unlock()

	var err error
	if !filepath.IsAbs(filename) {
		filename, err = inStandardDir(filename)
		if err != nil {
			return nil, err
		}
	}
	filename = renameExecutable(filename)

	if data != nil {
		log.Tracef("Placing executable in %s", filename)
		if err := filepersist.Save(filename, data, NewFileMode); err != nil {
			return nil, err
		}
		log.Trace("File saved, returning new Exec")
	} else {
		log.Tracef("Loading executable from %s", filename)
	}
	return newExec(filename)
}

// Command creates an exec.Cmd using the supplied args.
func (be *Exec) Command(args ...string) *exec.Cmd {
	return exec.Command(be.Filename, args...)
}

func newExec(filename string) (*Exec, error) {
	absolutePath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	return &Exec{Filename: absolutePath}, nil
}

func inStandardDir(filename string) (string, error) {
	folder, err := pathForRelativeFiles()
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(folder, NewFileMode)
	if err != nil {
		return "", fmt.Errorf("unable to make folder %s: %s", folder, err)
	}
	return filepath.Join(folder, filename), nil
}

func inHomeDir(filename string) (string, error) {
	log.Tracef("Determining user's home directory")
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("unable to determine user's home directory: %s", err)
	}
	return filepath.Join(usr.HomeDir, filename), nil
}
