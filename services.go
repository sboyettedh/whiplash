package whiplash

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// These are our Svc types, which are basically the types of ceph
// daemons.
const (
	MON = iota
	RGW
	OSD
)

var (
	// this is the list of admin socket commands we know
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
			s.Ping()
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
// TODO replace this with a standard aclient instance after 0.19
func (s *Svc) Query(req string) error {
	// make sure we know this command
	cmd, ok := cephcmds[req]
	if !ok {
		return fmt.Errorf("unknown request '%v'\n", req)
	}

	// make the connection
	conn, err := net.Dial("unix", s.Sock)
	if err != nil {
		return fmt.Errorf("could not connect to sock %v: %v\n", s.Sock, err)
	}
	defer conn.Close()

	// send command to the admin socket
	conn.SetDeadline(time.Now().Add(250 * time.Millisecond))
	_, err = conn.Write(cmd)
	if err != nil {
		return fmt.Errorf("could not write to %v: %v\n", s.Sock, err)
	}

	// zero our byte-collectors and bytes-read counter
	s.b1 = make([]byte, 64)
	s.b2 = s.b2[:0]
	s.mread = 0

	// get the response message length
	conn.SetDeadline(time.Now().Add(250 * time.Millisecond))
	n, err := conn.Read(s.b0)
	if err != nil {
		return fmt.Errorf("could not read message length on %v: %v\n", s.Sock, err)
	}
	if  n != 4 {
		return fmt.Errorf("too few bytes (%v) in message length on %v: %v\n", n, s.Sock, err)
	}
	buf := bytes.NewReader(s.b0)
	err = binary.Read(buf, binary.BigEndian, &s.mlen)
	if err != nil {
		return fmt.Errorf("could not decode message length on %v: %v\n", s.Sock, err)
	}

	// and read the message
	for {
		if s.mread == s.mlen {
			break
		}
		if x := s.mlen - s.mread; x < 64 {
			s.b1 = make([]byte, x)
		}
		conn.SetDeadline(time.Now().Add(250 * time.Millisecond))
		n, err := conn.Read(s.b1)
		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("could not read from %v: %v\n", s.Sock, err)
		}
		s.mread += int32(n)
		s.b2 = append(s.b2, s.b1[:n]...)
	}
	s.Resp = s.b2[:s.mlen]
	return err
}
