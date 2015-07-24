package elf

import (
	"testing"
)

func TestReadHeader(t *testing.T) {
	h, err := ReadHeaderInfo("<your test file here>")
	if err != nil {
		t.Error(err)
	}

	t.Log(h)
}
