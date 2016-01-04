package whiplash

import (
//	"net"
//	"time"
//	"bytes"
//	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"firepear.net/aclient"
)

// These are our Svc types, which are basically the types of ceph
// daemons.
const (
	MON = iota
	RGW
	OSD
)

var (
	// the list of admin socket commands we know
	cephcmds = map[string][]byte{"version": []byte("{\"prefix\":\"version\"}\000")}
)

// SvcCore
type SvcCore struct {
	// Name is the name/ID of the service
	Name string `json:"name"`

	// Type is the service/daemon type: MON, RGW, OSD
	Type int `json:"type"`

	// Host is the machine where the service runs
	Host string `json:"host"`

	// Version is the Ceph version of the service.
	Version string `json:"version"`

	// Reporting shows if a service is contactable and responsive
	Reporting bool `json:"reporting"`
}

// Svc represents a Ceph service
type Svc struct {
	Core *SvcCore

	// Sock is the admin socket for the service
	Sock string

	// Err holds the error (if any) from the Ping() check
	Err error

	// Resp receives response data from Query()
	Resp []byte

	// configuration for connections to the admin socket
	acconf *aclient.Config
	// b0 is where we read the message length into
	b0 []byte
	// mlen is the unpacked length from b0
	mlen int32
	// mread is the number of bytes read in the message so far
	mread int32
	// b1 is the buffer we read into from the network
	b1 []byte
	// b2 accumulates data from b1
	b2 []byte
}

type cephVersion struct {
	Version string `json:"version"`
}

// getCephServices examines wlc.CephConf and populates wlc.Svcs
func (wlc *WLConfig) getCephServices() {
	wlc.Svcs = make(map[string]*Svc)
	// iterate over CephConf, adding OSDs and RGWs
	for k, m := range wlc.CephConf {
		s := &Svc{Core: &SvcCore{Name: k}, b0: make([]byte, 4)}
		switch {
		case strings.HasPrefix(k, "osd."):
			s.Core.Type = OSD
			s.Core.Host = m["host"]
			s.Sock = strings.Replace(wlc.CephConf["osd"]["admin socket"], "$name", k, 1)
		case strings.HasPrefix(k, "client.radosgw"):
			s.Core.Type = RGW
			s.Core.Host = os.Getenv("HOSTNAME")
			if rsp, ok := m["rgw socket path"]; ok {
				s.Sock = rsp
			} else {
				s.Sock = strings.Replace(m["admin socket"], "$name", k, 1)
			}
		case strings.HasPrefix(k, "mon." + os.Getenv("HOSTNAME")):
			s.Core.Type = MON
			s.Core.Host = wlc.CephConf[k]["host"]
			s.Sock = strings.Replace(wlc.CephConf["osd"]["admin socket"], "$name", k, 1)
		}
		// only add defined services to Svcs when the admin socket exists
		if _, err := os.Stat(s.Sock); err == nil {
			s.acconf = &aclient.Config{Addr: s.Sock, Timeout: 100, NoPrefix: true}
			wlc.Svcs[k] = s
		}
	}
}

// Ping sends a version request to a Ceph service. It acts as the test
// for whether a service is reporting. When successful, it sets
// Reporting to 'true' and sets the service's Version. When it fails,
// Reporting is set to 'false', and Err is set to the returned error.
func (s *Svc) Ping() {
	err := s.Query("version")
	if err != nil {
		s.Core.Reporting = false
		s.Err = err
		return
	}
	var vs cephVersion
	err = json.Unmarshal(s.Resp, &vs)
	if err != nil {
		s.Core.Reporting = false
		s.Err = err
		return
	}
	s.Core.Reporting = true
	s.Err = nil
	s.Core.Version = vs.Version
}

// Query sends a request to a Ceph service and reads the result.
func (s *Svc) Query(req string) error {
	// make sure we know this command
	cmd, ok := cephcmds[req]
	if !ok {
		return fmt.Errorf("unknown request '%v'\n", req)
	}

	// make the connection
	c, err := aclient.NewUnix(*s.acconf)
	if err != nil {
		return fmt.Errorf("could not connect to sock %s: %s\n", s.Sock, err)
	}
	defer c.Close()

	// dispatch and return
	s.Resp, err = c.Dispatch(cmd)
	if err != nil {
		return fmt.Errorf("could not read reply on %s: %s\n", s.Sock, err)
	}
	return err
}
