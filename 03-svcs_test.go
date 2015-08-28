package whiplash

import (
	"os"
	"testing"
)

func TestGetOSDSvcs(t *testing.T) {
	// read the OSD config file
	c, err := New("./test_corpus/testosd.config")
	if err != nil {
		t.Fatalf("Tried using testosd.config file, but got: %v", err)
	}
	// do we have services?
	if len(c.Svcs) != 3 {
		t.Errorf("Svcs should be len 3 but is: %v", len(c.Svcs))
	}
	// are they as we expect them to be?
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

func TestGetRGWSvcs(t *testing.T) {
	os.Setenv("HOSTNAME", "peon9999")
	// read the RGW config file
	c, err := New("./test_corpus/testrgw.config")
	if err != nil {
		t.Fatalf("Tried using testrgw.config file, but got: %v", err)
	}
	// do we have services?
	if len(c.Svcs) != 1 {
		t.Errorf("Svcs should be len 1 but is: %v", len(c.Svcs))
	}
	// are they as we expect them to be?
	for _, svcname := range []string{"client.radosgw.peon9999"} {
		svc, ok := c.Svcs[svcname]
		if !ok {
			t.Fatalf("Should have found svc %v but did not", svcname)
		}
		// check type
		if svc.Type != RGW {
			t.Errorf("%v should have type RGW but is: '%v'", svcname, svc.Type)
		}
		// and assigned host
		if svc.Host != "peon9999" {
			t.Errorf("%v should be 'peon9999' but is: '%v'", svcname, svc.Type)
		}
		// and admin socket path
		svcsock := "/var/run/ceph/radosgw.client.radosgw.peon9999"
		if svc.Sock != svcsock {
			t.Errorf("%v should have sock /var/run/ceph/radosgw.client.radosgw.peon9999 but is: '%v'", svcname, svcsock, svc.Sock)
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
