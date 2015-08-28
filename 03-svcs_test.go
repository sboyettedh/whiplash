package whiplash

import (
	"testing"
)

func TestGetSvcs(t *testing.T) {
	// read the test config file
	c, err := New("./test_corpus/testosd.config")
	if err != nil {
		t.Fatalf("Tried using testosd.config file, but got: %v", err)
	}
	// do we have services?
	if len(c.Svcs) == 0 {
		t.Fatalf("Svcs should contian elements but is zero length")
	}
}
