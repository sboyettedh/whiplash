package whiplash // github.com/sboyettedh/whiplash

const (
	Version = "0.1.0"
)

// Row represents a datacenter row. It contains Racks.
type Row struct {
	ID int
	Name string
	Children []int
}

// Rack represents a datacenter rack. It contains Hosts.
type Rack struct {
	ID int
	Name string
	Children []int
}

// Host represents a machine running Ceph services. Svcs belong to it.
type Host struct {
	ID int
	Name string
	Rack string
}

// Request is the struct used for interchange between whiplash clients
// and the aggregator. Each network request consists of the request
// name followed by whitespace followed by a JSON-encoded Request.
type Request struct {
	// Svc is the core identifying and status info about the service
	// making the request.
	Svc *SvcCore `json:"svc"`

	// Payload is the data accompanying the request. May be empty, as
	// in a ping request.
	Payload interface{} `json:"payload"`
}

// OSD holds data specific to OSD services.
type OSD struct {
	// Weight is the crush weight of the OSD.
	Weight float32
	// Cap is the percentage of OSD storage capacity used.
	Cap float32
}
