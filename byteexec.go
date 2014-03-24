// Package byteexec provides a very basic facility for running executables
// supplied as byte arrays.
package byteexec

import (
	"io/ioutil"
	"os"
	"os/exec"
)

type ByteExec struct {
	tmpFile *os.File
}

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
	tmpFile.Chmod(0755)
	be = &ByteExec{tmpFile: tmpFile}
	return
}

func (be *ByteExec) Command(args ...string) *exec.Cmd {
	return exec.Command(be.tmpFile.Name(), args...)
}

func (be *ByteExec) Close() error {
	if be.tmpFile == nil {
		return nil
	} else {
		err1 := be.tmpFile.Close()
		err2 := os.Remove(be.tmpFile.Name())
		if err2 != nil {
			return err2
		} else {
			return err1
		}
	}
}
