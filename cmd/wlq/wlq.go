package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"firepear.net/aclient"
	"firepear.net/gaot"
	"github.com/sboyettedh/whiplash"
)

var (
	whipconf string
	dumpjson bool
	commands = gaot.NewFromString("status", nil)
)

func init() {
	// setup options vars
	flag.StringVar(&whipconf, "whipconf", "/etc/whiplash.conf", "Whiplash configuration file")
	flag.BoolVar(&dumpjson, "j", false, "Output query response as raw JSON")
	// load up command structure into trie
	cmdtail := commands.FindString("status")
	cmdtail.InsertString("cluster", nil)
}

func main() {
	// read whiplash config
	flag.Parse()
	// genconf should be FALSE in the whiplash.New call for a wlq
	// instance: we don't expect to hacve ceph services around on a
	// machine running a query.
	wl, err := whiplash.New(whipconf, false)
	if err != nil {
		log.Fatalf("error reading configuration file: %s\n", err)
	}

	// validate user input
	err = validateInput()

	// set up configuration and create aclient instance
	acconf := aclient.Config{
		Addr: wl.Aggregator.BindAddr + ":" + wl.Aggregator.QueryPort,
		Timeout: 100,
	}
	c, err := aclient.NewTCP(acconf)
	if err != nil {
		log.Fatalf("error creating network connection: %s", err)
	}
	defer c.Close()

	// stitch together the non-option arguments into our request
	req := strings.Join(flag.Args(), " ")
	// and dispatch it to the server!
	respj, err := c.Dispatch([]byte(req))
	if err != nil {
		log.Fatalf("error creating network connection: %s", err)
	}

	// vivify response and handle errors
	resp := new(whiplash.QueryResponse)
	err = json.Unmarshal(respj, &resp)
	if err != nil {
		log.Fatalf("error unmarshaling json\nresponse: '%s'", string(respj))
	}
	if resp.Code >= 400 {
		log.Fatalf("there was a problem with the request:\n%s", string(respj))
	}

	// if -j has been specified, print the raw response data and exit
	if dumpjson {
		fmt.Println(string(resp.Data))
		os.Exit(0)
	}

	// else, we have to hand off to a pretty-printing routine
	// TODO write a pretty-printing routine
}

func validateInput() error {
	// get remaining arguments
//	args := flag.Args()
	// slam together remaining args
//	req := strings.Join(flag.Args(), "")
	return nil
}
