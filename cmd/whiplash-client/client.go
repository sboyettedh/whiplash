package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"firepear.net/aclient"
	"github.com/sboyettedh/whiplash"
)

var (
	whipconf string
	hostname string
)

func init() {
	flag.StringVar(&whipconf, "whipconf", "/etc/whiplash.json", "Whiplash configuration file")
	hostname, _ = os.LookupEnv("HOSTNAME")
}

func main() {
	flag.Parse()
	wl, err := whiplash.New(whipconf, true)
	if err != nil {
		log.Printf("%v: could not read configuration file: %v\n", os.Args[0], err)
		os.Exit(1)
	}

	// we work with Svcs; these are for reporting
	osds := make(map[string]*whiplash.Osd)

	// ask known services what version of Ceph they're running
	for svcname, svc := range wl.Svcs {
		if svc.Reporting == false {
			log.Printf("%v: not reporting: %v", svcname, svc.Err)
			continue
		}
		switch svc.Type {
		case whiplash.OSD:
			var osderr string
			if svc.Err == nil {
				osderr = ""
			} else {
				osderr = svc.Err.Error()
			}
			osds[svcname] = &whiplash.Osd{
				Host: hostname,
				Version: svc.Version,
				Reporting: svc.Reporting,
				Err: osderr,
			}
		}
	}

	// set up the aclient configuration
	acconf := aclient.Config{
		Addr: wl.Aggregator.BindAddr + ":" + wl.Aggregator.BindPort,
		Timeout: 100,
	}
	// send our data
	ac, err := aclient.NewTCP(acconf)
	if err != nil {
		panic(err)
	}
	defer ac.Close()
	josd, err := json.Marshal(osds)
	if err != nil {
		panic(err)
	}
	req := []byte("osdupdate ")
	req = append(req, josd...)
	resp, err := ac.Dispatch(req)
	if err != nil {
		panic(err)
	}
	log.Println(string(resp))
}
