package whiplash

// these are our ceph service types
const (
	MON = iota
	RGW
	OSD
)

type Svc struct {
	// Type is the service type, as enumerated above
	Type int

	// Sock is the admin socket for the service
	Sock string
}
