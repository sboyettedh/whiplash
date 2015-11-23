package whiplash // github.com/sboyettedh/whiplash

const (
	Version = "0.1.0"
)

type Row struct {
	ID int
	Name string
	Children []int
}

type Rack struct {
	ID int
	Name string
	Children []int
}

type Host struct {
	ID int
	Name string
	Rack string
	Weight float
}

type Osd struct {
	Host string
	Version string
	Reporting bool
	Capacity int
	Utilized int
	Err string
}
