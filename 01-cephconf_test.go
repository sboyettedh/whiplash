package whiplash

import (
	"testing"
)

func TestCephMonConf(t *testing.T) {
	// bad file first, just to be sure
	cephconf, err := parseCephConf("./zzzxxx")
	if err == nil {
		t.Errorf("Opening nonexistent file worked; got: %v", cephconf)
	}
	// now with good file
	cephconf, err = parseCephConf("./test_corpus/ceph.mon.conf")
	if err != nil {
		t.Errorf("Opening ceph.mon.conf failed; got: %v", err)
	}
	// check for correct number of sections
	if len(cephconf) != 8 {
		t.Errorf("Expected 8 sections but got %v", len(cephconf))
	}
	// and the section names
	keys := []string{"global", "mon", "mon.peon9995", "mon.peon9996", "mon.peon9997",
		             "mon.peon9998", "mon.peon9999", "osd"}
	for _, key := range keys {
		if _, ok := cephconf[key]; !ok {
			t.Errorf("Expected `%v` to be a section, but it is not", key)
		}
	}
	// then check an actual value to make sure it's all working
	if cephconf["osd"]["admin socket"] != "/var/run/ceph/ceph-$name.asok" {
		t.Errorf("`admin socket` value not as expected: %v", cephconf["osd"]["admin socket"])
	}
}

func TestCephOsdConf(t *testing.T) {
	cephconf, err := parseCephConf("./test_corpus/ceph.osd.conf")
	if err != nil {
		t.Errorf("Opening ceph.osd.conf failed; got: %v", err)
	}
	// check for correct number of sections
	if len(cephconf) != 12 {
		t.Errorf("Expected 12 sections but got %v", len(cephconf))
	}
	// and the section names
	keys := []string{"global", "mon", "mon.peon9995", "mon.peon9996", "mon.peon9997",
		"mon.peon9998", "mon.peon9999", "osd", "osd.9900", "osd.9901", "osd.9902"}
	for _, key := range keys {
		if _, ok := cephconf[key]; !ok {
			t.Errorf("Expected `%v` to be a section, but it is not", key)
		}
	}
	// then check an actual value to make sure it's all working
	if cephconf["osd.9901"]["host"] != "cephstore9999" {
		t.Errorf("`host` value not as expected: %v", cephconf["osd.9901"]["host"])
	}
}

func TestCephRgwConf(t *testing.T) {
	cephconf, err := parseCephConf("./test_corpus/ceph.rgw.conf")
	if err != nil {
		t.Errorf("Opening ceph.osd.conf failed; got: %v", err)
	}
	// check for correct number of sections
	if len(cephconf) != 10 {
		t.Errorf("Expected 10 sections but got %v", len(cephconf))
	}
	// and the section names
	keys := []string{"global", "mon", "mon.peon9995", "mon.peon9996", "mon.peon9997",
		"mon.peon9998", "mon.peon9999", "osd", "client.radosgw.peon9999"}
	for _, key := range keys {
		if _, ok := cephconf[key]; !ok {
			t.Errorf("Expected `%v` to be a section, but it is not", key)
		}
	}
	// then check an actual value to make sure it's all working
	if cephconf["client.radosgw.peon9999"]["rgw dns name"] != "objects.example.com" {
		t.Errorf("`rgw dns name` value not as expected: %v", cephconf["client.radosgw.peon9999"]["rgw dns name"])
	}
}

