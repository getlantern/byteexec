package byteexec

import (
	"testing"
)

func TestExec(t *testing.T) {
	bytes, err := Asset("upnpc")
	if err != nil {
		t.Fatalf("Unable to read upnpc program: %s", err)
	}
	be, err := NewByteExec(bytes)
	if err != nil {
		t.Fatalf("Unable to create new ByteExec: %s", err)
	}
	defer be.Close()
	cmd := be.Command("-s")
	err = cmd.Run()
	if err != nil {
		t.Errorf("Unable to run command: %s", err)
	}
}
