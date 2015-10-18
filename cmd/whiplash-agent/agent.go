package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sboyettedh/whiplash"
)

var (
	whipconf string
)

func init() {
	flag.StringVar(&whipconf, "whipconf", "/etc/whiplash.json", "Whiplash configuration file")
}

func main() {
	flag.Parse()
	wl, err := whiplash.New(whipconf, true)
	if err != nil {
		fmt.Printf("%v: could not read configuration file: %v\n", os.Args[0], err)
		os.Exit(1)
	}

	// ask known services what version of Ceph they're running
	for svcname, svc := range wl.Svcs {
		if svc.Reporting == false {
			fmt.Printf("%v: not reporting: %v", svcname, svc.Err)
			continue
		}
		fmt.Printf("%v: %q\n", svcname, svc.Version)
	}
}
