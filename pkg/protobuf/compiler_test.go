package protobuf

import (
	"testing"
)

func TestNewSystemCompiler(t *testing.T) {
	if compiler, err := NewSystemCompiler(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("System compiler version: %s", compiler.Version)
	}
}
