package whiplash

// This file contains the code for parsing the output of 'ceph osd
// tree', which I don't want to do at all, but I haven't figured out
// where else to get the weight and reweight values of OSDs from.
//
// Other than not wanting to shell out and depend on the Python 'ceph'
// tool, there's the fact that the output of 'ceph osd tree -f json'
// is not regular. At the top level, it's this:
//
// {nodes: [], stray: []}
//
// ...which is easy enough. The problem is that there are two
// different things that "nodes" can contain. (And I don't know what a
// "stray" can be. I assume the same thing as a node, but detached
// from the CRUSH hierarchy somehow?)
//
// It seems that a node is either a container (row, rack, host), or an
// OSD. The former all share a collective structure:
//
// { "id": -4, "name": "irv-n", "type": "row", "type_id": 3, "children": []}
//
// But an OSD looks like this:
//
// {"id": 48, "name": "osd.48", "type": "osd", "type_id": 0, "crush_weight": 2.000000,
//  "depth": 4, "exists": 1, "status": "up", "reweight": 1.000000, "primary_affinity": 1.000000}
//
// Which would be fine, EXCEPT THAT there is no nesting in the JSON
// structure. Rows don't hold racks, which don't hold hosts, which
// don't hold OSDs. The "children" field of the container types,
// instead of holding actual node objects, or even the names of their
// child nodes, hold the ID numbers of their child nodes.
//
// To work around this, the top-level structure (osdtree) is defined
// as containing two slices of json.RawMessage. Then, when iterating
// over osdtree.Nodes, we take the most brute force approach possible:
// try to Unmarshal the RawMessage into an osdtreeContainer; if that
// produces an error, Unmarshal it into an osdtreeOSD.

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
)

// These three types 
type osdtree struct {
	Nodes []json.RawMessage `json:"nodes"`
	Stray []json.RawMessage `json:"stray"`
}
type osdtreeContainer struct {
	ID int         `json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type"`
	Type_ID int    `json:"type_id"`
	Children []int `json:"children"`
}
type osdtreeOSD struct {
	ID int           `json:"id"`
	Name string      `json:"name"`
	Type string     `json:"type"`
	Type_id int      `json:"type_id"`
	Weight float64   `json:"crush_weight"`
	Depth int        `json:"depth"`
	Exists int       `json:"exists"`
	Status string    `json:"status"`
	Reweight float64 `json:"reweight"`
	Affinity float64 `json:"primary_affinity"`
}

// osdtreeDump calls 'ceph osd tree -f json' and writes the output to
// a file.
//
// This is less efficient than other mechanisms, but it makes
// osdtreeParse testable.
func osdtreeDump(output string) error {
	jtree, err := exec.Command("ceph", "osd", "tree", "-f", "json").Output()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(output, jtree, 0644)
	return err
}

// osdtreeParse reads a JSON dump of 'ceph osd tree' and
// populates/updates datastructures with it.
func osdtreeParse(input string) error {
	jtree, err := ioutil.ReadFile(input)
	if err != nil {
		return err
	}
	var tree osdtree
	err = json.Unmarshal(jtree, &tree)
	if err != nil {
		return err
	}
	var jcont osdtreeContainer
	var josd osdtreeOSD
	for _, node := range tree.Nodes {
		err = json.Unmarshal(node, &josd)
		if err == nil {
			//fmt.Printf("Node %d: type %s, name %s\n", i, josd.Type, josd.Name)
		} else {
			json.Unmarshal(node, &jcont)
			//fmt.Printf("Node %d: type %s, name %s\n", i, jcont.Type, jcont.Name)
		}
	}
	return err
}
