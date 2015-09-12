package whiplash

import (
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

// Svc represents a Ceph service
type Svc struct {
	// Type is the service/daemon type: MON, RGW, OSD
	Type int

	// Sock is the admin socket for the service
	Sock string

	// Reporting shows if a service is contactable and responsive
	Reporting bool

	// Host is the machine where the service runs
	Host string

	// Version is the Ceph version of the service (if reporting)
	Version string

	// b1 is the buffer we read into from the network
	b1 []byte
	// b2 accumulates data from b1
	b2 []byte
}

type cephVersion struct {
	version string
}

// getCephServices examines wlc.CephConf and populates wlc.Svcs
func (wlc *WLConfig) getCephServices() {
	wlc.Svcs = make(map[string]*Svc)
	// iterate over CephConf, adding OSDs and RGWs
	for k, m := range wlc.CephConf {
		s := &Svc{b1: make([]byte, 64)}
		switch {
		case strings.HasPrefix(k, "osd."):
			s.Type = OSD
			s.Host = m["host"]
			s.Sock = strings.Replace(wlc.CephConf["osd"]["admin socket"], "$name", k, 1)
			err := s.Query("version")
			if err == nil {
				s.Reporting = true
			}
			vs := &cephVersion{}
			err = json.Unmarshal(s.b2, vs)
			if err == nil {
				s.Version = vs.version
			}
			wlc.Svcs[k] = s
		case strings.HasPrefix(k, "client.radosgw"):
			s.Type = RGW
			s.Host = os.Getenv("HOSTNAME")
			if rsp, ok := m["rgw socket path"]; ok {
				s.Sock = rsp
			} else {
				s.Sock = strings.Replace(m["admin socket"], "$name", k, 1)
			}
			err := s.Query("version")
			if err == nil {
				s.Reporting = true
			}
			vs := &cephVersion{}
			err = json.Unmarshal(s.b2, vs)
			if err == nil {
				s.Version = vs.version
			}
			wlc.Svcs[k] = s
		}
	}
	// if we get down here and Svcs is empty, we're on a monitor
	if len(wlc.Svcs) == 0 {
		k := "mon." + os.Getenv("HOSTNAME")
		s := &Svc{Type: MON, Host: wlc.CephConf[k]["host"], b1: make([]byte, 64)}
		s.Sock = strings.Replace(wlc.CephConf["osd"]["admin socket"], "$name", k, 1)
		err := s.Query("version")
		if err == nil {
			s.Reporting = true
		}
		vs := &cephVersion{}
		err = json.Unmarshal(s.b2, vs)
		if err == nil {
			s.Version = vs.version
		}
		wlc.Svcs[k] = s
	}
}

func (s *Svc) Query(cmd string) error {
	// make the connection
	conn, err := net.Dial("unix", s.Sock)
	if err != nil {
		return fmt.Errorf("could not connect to sock %v: %v\n", s.Sock, err)
	}
	defer conn.Close()

	// send command to the admin socket
	conn.SetDeadline(time.Now().Add(250 * time.Millisecond))
	_, err = conn.Write([]byte(cmd + "\000"))
	if err != nil {
		return fmt.Errorf("could not write to %v: %v\n", s.Sock, err)
	}

	// zero our byte-collector and read what we got back.
	s.b2 = s.b2[:0]
	conn.SetDeadline(time.Now().Add(50 * time.Millisecond))
	for {
		n, err := conn.Read(s.b1)
		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("could not read from %v: %v\n", s.Sock, err)
		}
		// since the admin-daemon closes the connection as soon as
		// it's done writing, there's no EOM to watch for. you just
		// read until there's nothing left, and then you're done.
		if n == 0 {
			break
		}
		s.b2 = append(s.b2, s.b1[:n]...)
	}
	return err
}
