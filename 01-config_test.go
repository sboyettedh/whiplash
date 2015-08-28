package whiplash

import (
	"testing"
)

func TestGetConfig(t *testing.T) {
	// try nonexistant file
	c, err := New("./test_corpus/zzzzxxxx")
	if err == nil {
		t.Errorf("Tried using nonexistant config file, but err is nil and config is %v", c)
	}
	// try non-json file
	c, err = New("./test_corpus/notjson.config")
	if err == nil {
		t.Errorf("Tried using non-json config file, but err is nil and config is %v", c)
	}
	// try json file which isn't a config for us
	c, err = New("./test_corpus/badjson.config")
	if err == nil {
		t.Errorf("Tried using non-json config file, but err is nil and config is %v", c)
	}
	// try json file which is a config but doesn't have an agg addr
	c, err = New("./test_corpus/badjson2.config")
	if err == nil {
		t.Errorf("Tried using bad config file, but err is nil and config is %v", c)
	}
	// use good config file
	c, err = New("./test_corpus/test.config")
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
	c, err := New("./test_corpus/test.config")
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
	c, err := New("./test_corpus/test2.config")
	if err == nil {
		t.Errorf("Opening nonexistent file worked; got: %v", c)
	}
	// read the test config file
	c, err = New("./test_corpus/test.config")
	if err != nil {
		t.Fatalf("Tried using test.config file, but got: %v", err)
	}
	if c.CephConf["osd"]["admin socket"] != "/var/run/ceph/ceph-$name.asok" {
		t.Errorf("`admin socket` value not as expected: %v", c.CephConf["osd"]["admin socket"])
	}
}
