package whiplash

type Node struct {
	Type int
	Name string
	Rack string
}

type Osd struct {
	Host string
	Version string
	Reporting bool
	Err string
}
