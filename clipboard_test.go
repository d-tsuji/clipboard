package clipboard

import (
	"testing"
)

func TestCopyAndPaste(t *testing.T) {
	want := "gopher"
	if err := Set(want); err != nil {
		t.Error(err)
	}

	got, err := Get()
	if err != nil {
		t.Error(err)
	}

	if got != want {
		t.Errorf("CopyAndPaste mismatch: got %v, want %v", got, want)
	}
}
