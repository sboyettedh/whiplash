package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sboyettedh/whiplash"
)

// these are our ceph service types
const (
	MON = iota
	RGW
	OSD
)

var (
	whipconf string
	hostname = os.Getenv("HOSTNAME")
	svcs map[string]cephsvc
)

type cephsvc struct {
	// stype is the service type, as enumerated above
	stype int
	sock string
}

func init() {
	flag.StringVar(&whipconf, "whipconf", "/etc/whiplash.json", "Whiplash configuration file")
}

func main() {
	flag.Parse()
	wl, err := whiplash.New(whipconf)
	if err != nil {
		fmt.Printf("%v: could not read configuration file: %v\n", os.Args[0], err)
		os.Exit(1)
	}

	// hit the OSD admin sockets we know about and ask for ceph version
	for key, _ := range wl.CephConf {
		if strings.HasPrefix(key, "osd.") {
			sock := strings.Replace(wl.CephConf["osd"]["admin socket"], "$name", key, 1)
			fmt.Printf("%v ---------------------------\n", sock)
			res, err := askceph(sock, "{\"prefix\":\"version\"}")
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("%v\n", string(res))
		}
	}
}
