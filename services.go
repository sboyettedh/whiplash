package whiplash

// These are our Svc types, which are basically the types of ceph
// daemons.
const (
	MON = iota
	RGW
	OSD
)

// Svc represents a Ceph service
type Svc struct {
	// Type is the service type: MON, RGW, OSD
	Type int

	// Sock is the admin socket for the service
	Sock string

	// Reporting shows if a service is contactable and responsive
	Reporting bool
}

