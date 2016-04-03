package whiplash

import (
	"testing"
)

func TestGetIface(t *testing.T) {
	// try to match nothing
	a, err := getIface("foo")
	if a != nil {
		t.Errorf("'foo' should not have matched an interface, but got: %v", a)
	}
	if err.Error() != "No interfaces match 'foo'" {
		t.Errorf("Expected `No interfaces match 'foo'` but got: %v", err)
	}
	// try to match more than one thing
	a, err = getIface(".")
	if a != nil {
		t.Errorf("'.' should have returned nil, but got: %v", a)
	}
	if err.Error() != "Multiple interfaces match '.'" {
		t.Errorf("Expected `Multiple interfaces match '.'` but got: %v", err)
	}
	// try to match localhost
	a, err = getIface("127.0.0")
	if err != nil {
		t.Errorf("Expected to match localhost but got: %v", err)
	}
	if a.String() != "127.0.0.1/8" {
		t.Errorf("Expected to match localhost but got: %v", a.String())
	}
	if a.Network() != "ip+net" {
		t.Errorf("Expected a TCP addr but got: %v", a.Network())
	}
}
