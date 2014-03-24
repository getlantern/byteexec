// Package byteexec provides a very basic facility for running executables
// supplied as byte arrays, which is handy when used with
// github.com/jteeuwen/go-bindata.
//
// ByteExec works by storing the provided command in a temp file.  A ByteExec
// should always be closed using its Close() method to clean up the temp file.
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
	"io/ioutil"
	"os"
	"os/exec"
)

type ByteExec struct {
	fileName string
}

// NewByteExec creates a new ByteExec using the program stored in the provided
// bytes.
func NewByteExec(bytes []byte) (be *ByteExec, err error) {
	var tmpFile *os.File
	tmpFile, err = ioutil.TempFile("", "byteexec_")
	if err != nil {
		return
	}
	_, err = tmpFile.Write(bytes)
	if err != nil {
		return
	}
	tmpFile.Sync()
	tmpFile.Chmod(0755)
	tmpFile.Close()

	orig := tmpFile.Name()
	renamed := renameExecutable(orig)
	if renamed != orig {
		err = os.Rename(orig, renamed)
		if err != nil {
			return nil, err
		}
	}
	be = &ByteExec{fileName: renamed}
	return
}

// Command creates an exec.Cmd using the supplied args.
func (be *ByteExec) Command(args ...string) *exec.Cmd {
	return exec.Command(be.fileName, args...)
}

// Close() closes the ByteExec, cleaning up the associated temp file.
func (be *ByteExec) Close() error {
	if be.fileName == "" {
		return nil
	} else {
		return os.Remove(be.fileName)
	}
}
