package whiplash

import (
	"encoding/json"
	"fmt"
	"os"
	"net"
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
}

type cephVersion struct {
	version string
}

// getCephServices examines wlc.CephConf and populates wlc.Svcs
func (wlc *WLConfig) getCephServices() {
	// iterate over CephConf, adding OSDs and RGWs
	for k, m := range wlc.CephConf {
		switch {
		case strings.HasPrefix(k, "osd."):
			s := &Svc{Type: OSD, Host: m["host"]}
			s.Sock = strings.Replace(m["admin socket"], "$name", k, 1)
			vbytes, err := adminSockQuery(s.Sock, "version")
			if err == nil {
				s.Reporting = true
			}
			vs := &cephVersion{}
			err = json.Unmarshal(vbytes, vs)
			if err == nil {
				s.Version = vs.version
			}
			wlc.Svcs[k] = s
		case strings.HasPrefix(k, "client.radosgw"):
			s := &Svc{Type: RGW, Host: m["host"]}
			if rsp, ok := m["rgw socket path"]; ok {
				s.Sock = rsp
			} else {
				s.Sock = strings.Replace(m["admin socket"], "$name", k, 1)
			}
			vbytes, err := adminSockQuery(s.Sock, "version")
			if err == nil {
				s.Reporting = true
			}
			vs := &cephVersion{}
			err = json.Unmarshal(vbytes, vs)
			if err == nil {
				s.Version = vs.version
			}
			wlc.Svcs[k] = s
		}
	}
	// if we get down here and Svcs is empty, we're on a monitor
	if wlc.Svcs == nil {
		k := "mon." + os.Getenv("HOSTNAME")
		s := &Svc{Type: MON, Host: wlc.CephConf[k]["host"]}
		s.Sock = strings.Replace(wlc.CephConf[k]["admin socket"], "$name", k, 1)
		vbytes, err := adminSockQuery(s.Sock, "version")
		if err == nil {
			s.Reporting = true
		}
		vs := &cephVersion{}
		err = json.Unmarshal(vbytes, vs)
		if err == nil {
			s.Version = vs.version
		}
		wlc.Svcs[k] = s
	}
}

func adminSockQuery(sock, cmd string) ([]byte, error) {
	b1 := make([]byte, 64)
	var b2 []byte

	// make the connection
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, fmt.Errorf("could not connect to sock %v: %v\n", sock, err)
	}
	defer conn.Close()

	// send command to the admin socket
	conn.SetDeadline(time.Now().Add(250 * time.Millisecond))
	_, err = conn.Write([]byte(cmd + "\000"))
	if err != nil {
		return nil, fmt.Errorf("could not write to %v: %v\n", sock, err)
	}

	// now read what we got back.
	conn.SetDeadline(time.Now().Add(50 * time.Millisecond))
	for {
		n, err := conn.Read(b1)
		if err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("could not read from %v: %v\n", sock, err)
		}
		// since the admin-daemon closes the connection as soon as
		// it's done writing, there's no EOM to watch for. you just
		// read until there's nothing left, and then you're done.
		if n == 0 {
			break
		}
		b2 = append(b2, b1[:n]...)
	}
	return b2, err
}
