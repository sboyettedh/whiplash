package whiplash // github.com/sboyettedh/whiplash

const (
	Version = "0.1.0"
)

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
