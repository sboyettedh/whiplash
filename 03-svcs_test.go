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
	if len(c.Svcs) != 3 {
		t.Errorf("Svcs should be len 3 but is: %v", len(c.Svcs))
	}
	// are they the ones we expect?
	for _, svcname := range []string{"osd.9900", "osd.9901", "osd.9902"} {
		svc, ok := c.Svcs[svcname]
		if !ok {
			t.Fatalf("Should have found svc %v but did not", svcname)
		}
		// check type
		if svc.Type != OSD {
			t.Errorf("%v should have type OSD but is: '%v'", svcname, svc.Type)
		}
		// and assigned host
		if svc.Host != "cephstore9999" {
			t.Errorf("%v should be 'cephstore9999' but is: '%v'", svcname, svc.Type)
		}
		// and admin socket path
		svcsock := "/var/run/ceph/ceph-" + svcname + ".asok"
		if svc.Sock != svcsock {
			t.Errorf("%v should have sock %v but is: '%v'", svcname, svcsock, svc.Sock)
		}
		// Reporting should be false, and we should not have a version string
		if svc.Reporting == true {
			t.Errorf("%v should not be reporting, but claims to be", svcname)
		}
		if svc.Version != "" {
			t.Errorf("%v should not have a version, but has '%v'", svcname, svc.Version)
		}
	}
}
