package whiplash

import (
	"testing"
)

func TestGetSvcs(t *testing.T) {
	// read the test config file
	c, err := New("./test_corpus/test.config")
	if err != nil {
		t.Fatalf("Tried using test.config file, but got: %v", err)
	}
}
