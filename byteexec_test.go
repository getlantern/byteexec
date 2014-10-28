package byteexec

import (
	"testing"

	"github.com/getlantern/testify/assert"
)

func TestExec(t *testing.T) {
	bytes, err := Asset("helloworld")
	if err != nil {
		t.Fatalf("Unable to read helloworld program: %s", err)
	}
	be, err := NewByteExec(bytes)
	if err != nil {
		t.Fatalf("Unable to create new ByteExec: %s", err)
	}
	defer be.Close()
	cmd := be.Command()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Unable to run helloworld program: %s", err)
	}

	assert.Equal(t, "Hello world\n", string(out), "Did not receive expected output from helloworld program")
}
