package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"firepear.net/aclient"
	"github.com/sboyettedh/whiplash"
)

var (
	whipconf string
)

func init() {
	flag.StringVar(&whipconf, "whipconf", "/etc/whiplash.conf", "Whiplash configuration file")
}

func main() {
	// read whiplash config
	flag.Parse()
	// genconf should be FALSE in the whiplash.New call for a wlq
	// instance: we don't expect to hacve ceph services around on a
	// machine running a query.
	wl, err := whiplash.New(whipconf, false)
	if err != nil {
		log.Fatalf("error reading configuration file: %v\n", err)
	}
	flag.Parse()

	// set up configuration and create aclient instance
	acconf := aclient.Config{
		Addr: wl.Aggregator.BindAddr + ":" + wl.Aggregator.QueryPort,
		Timeout: 100,
	}
	c, err := aclient.NewTCP(acconf)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// stitch together the non-option arguments into our request
	req := strings.Join(flag.Args(), " ")

	// and dispatch it to the server!
	resp, err := c.Dispatch([]byte(req))
	if err != nil {
		panic(err)
	}

	// print out what we got back and exit
	fmt.Println(string(resp))
}
