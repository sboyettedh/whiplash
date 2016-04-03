package whiplash

import (
	"testing"
)


func TestGetConfig(t *testing.T) {
	// try nonexistant file
	c, err := New("./test_corpus/zzzzxxxx", false)
	if err == nil {
		t.Errorf("Tried using nonexistant config file, but err is nil and config is %v", c)
	}
	// try non-json file
	c, err = New("./test_corpus/notjson.config", false)
	if err == nil {
		t.Errorf("Tried using non-json config file, but err is nil and config is %v", c)
	}
	// try json file which isn't a config for us
	c, err = New("./test_corpus/badjson.config", false)
	if err == nil {
		t.Errorf("Tried using non-json config file, but err is nil and config is %v", c)
	}
	// try json file which is a config but doesn't have an agg addr
	c, err = New("./test_corpus/badjson2.config", false)
	if err == nil {
		t.Errorf("Tried using bad config file, but err is nil and config is %v", c)
	}
	// use good config file
	c, err = New("./test_corpus/test.config", false)
	if err != nil {
		t.Fatalf("Tried using test.config file, but got: %v", err)
	}
	if c.CephConfLoc != "./test_corpus/ceph.mon.conf" {
		t.Errorf("c.CephConfLoc should be ./test_corpus/ceph.mon.conf but got: %v", c.CephConfLoc)
	}
	if c.Aggregator.BindAddr != "127.0.0.1" {
		t.Errorf("c.Aggregator.BindAddr should be 127.0.0.1 but got: %v", c.Aggregator.BindAddr)
	}
	if c.Aggregator.BindPort != "61089" {
		t.Errorf("c.Aggregator.BindPort should be 61089 but got: %v", c.Aggregator.BindPort)
	}
}

func TestGetConfigGetIface(t *testing.T) {
	// read the test config file
	c, err := New("./test_corpus/test.config", false)
	if err != nil {
		t.Fatalf("Tried using test.config file, but got: %v", err)
	}
	_, err = getIface(c.Aggregator.BindAddr)
	if err != nil {
		t.Fatalf("Expected 127.0.0.1 to be a valid iface but got: %v", err)
	}
}

func TestGetConfigParseCephConf(t *testing.T) {
	// bad file first, just to be sure
	c, err := New("./test_corpus/nofilenamedthis.config", true)
	if err == nil {
		t.Errorf("Opening nonexistent file worked; got: %v", c)
	}
	// read the test config file
	c, err = New("./test_corpus/test.config", true)
	if err != nil {
		t.Fatalf("Tried using test.config file, but got: %v", err)
	}
	if c.CephConf["osd"]["admin socket"] != "./test_corpus/monsocks/ceph-$name.asok" {
		t.Errorf("`admin socket` value not as expected: '%v'", c.CephConf["osd"]["admin socket"])
	}
}

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
	if cephconf["osd"]["admin socket"] != "./test_corpus/monsocks/ceph-$name.asok" {
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
