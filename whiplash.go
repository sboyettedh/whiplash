package whiplash // github.com/sboyettedh/whiplash

import (
	"encoding/json"
)

const (
	Version = "0.3.0"
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

// ClientRequest is the struct used for interchange between whiplash clients
// and the aggregator. Each network request consists of the request
// name followed by whitespace followed by a JSON-encoded Request.
type ClientRequest struct {
	// Svc is the core identifying and status info about the service
	// making the request.
	Svc *SvcCore `json:"svc"`

	// Payload is the data accompanying the request. May be empty, as
	// in a ping request.
	Payload json.RawMessage `json:"payload"`
}

// OSDsvc holds data specific to OSD services.
type OSDsvc struct {
	// Weight is the crush weight of the OSD.
	Weight float32
	// Cap is the percentage of OSD storage capacity used.
	Cap float32
}

type QueryResponse struct {
	Code int `json:"code"`
	Cmd string `json:"cmd"`
	Subcmd string `json:"subcmd"`
	Args []string `json:"args"`
	Data json.RawMessage `json:"data"`
}
