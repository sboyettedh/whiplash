package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"os"
	"time"

	"firepear.net/pclient"
	"github.com/sboyettedh/whiplash"
)

var (
	whipconf = flag.String("c", "/etc/whiplash.conf", "Whiplash configuration file")
	hostname, _ = os.LookupEnv("HOSTNAME")
	pcconf *pclient.Config
	// nil payload for pings
	nilPayload, _ = json.Marshal(nil)
	// which interval set to use for tickers
	intv int
	// the interval sets
	intvs = []map[string]time.Duration{
		{"ping": 11, "stat": 313},
		{"ping": 13, "stat": 311},
		{"ping": 17, "stat": 307},
		{"ping": 19, "stat": 293},
	}
)

func main() {
	flag.Parse()
	wl, err := whiplash.New(*whipconf, true)
	if err != nil {
		log.Fatalf("error reading configuration file: %v\n", err)
	}
	sigchan := whiplash.AppSetup("whiplash-client", "0.1.0", pclient.Pkgname, pclient.Version)
	defer whiplash.AppCleanup("whiplash-client")

	// need a pclient configuration to talk to the aggregator with
	pcconf = &pclient.Config{
		Addr: wl.Aggregator.BindAddr + ":" + wl.Aggregator.BindPort,
		Timeout: wl.Client.Timeout,
	}

	// decide what notification interval to use
	rand.Seed(time.Now().UnixNano())
	intv = rand.Intn(len(intvs))
	log.Printf("using interval set: %q\n", intvs[intv])
	// create tickers and launch monitor funcs
	pingticker := time.NewTicker(time.Second * intvs[intv]["ping"])
	statticker := time.NewTicker(time.Second * intvs[intv]["stat"])
	go pingSvcs(wl.Svcs, pingticker.C)
	go statSvcs(wl.Svcs, statticker.C)

	// mainloop
	keepalive := true
	for keepalive {
		select {
		case <-sigchan:
			// we've trapped a signal from the OS. tell our Asock to
			// shut down, but don't exit the eventloop because we want
			// to handle the Msgs which will be incoming.
			log.Println("OS signal received; shutting down")
			keepalive = false
		}
		// there's no default case in the select, as that would cause
		// it to be nonblocking. and that would cause main() to exit
		// immediately.
	}
}


// pingScvs, running on the ticker channel `tc`, polls each service on
// this node. this is a basic check for being alive, based on doing a
// version request. it then reports to the aggregator.
func pingSvcs(svcs map[string]*whiplash.Svc, tc <-chan time.Time) {
	// every tick...
	for _ = range tc {
		// loop over known services
		log.Println("updating: ping")
		for _, svc := range svcs {
			// ping, then send data if there are no issues
			svc.Ping()
			if svc.Err != nil {
				log.Println("ping failed:", svc.Core.Name, svc.Err)
				continue
			}
			sendData("ping", svc.Core.Name, &whiplash.ClientUpdate{
				Time: time.Now().Unix(),
				Svc: svc.Core,
				Payload: nilPayload,
			})
		}
	}
}

// statSvcs performs a more in-depth status check on services. it
// operates nearly identically to pingSvcs.
func statSvcs(svcs map[string]*whiplash.Svc, tc <-chan time.Time) {
	var statdata json.RawMessage
	for _ = range tc {
		log.Println("updating: stat")
		for _, svc := range svcs {
			statdata = svc.Stat()
			// TODO handle MONs and RGWs
			if svc.Err != nil {
				log.Println("stat failed:", svc.Core.Name, svc.Err)
				continue
			}
			sendData("stat", svc.Core.Name, &whiplash.ClientUpdate{
				Time: time.Now().Unix(),
				Svc: svc.Core,
				Payload: statdata,
			})
		}
	}
}

// sendData handles the actual sending of data to the aggregator.
func sendData(cmd, svc string, u *whiplash.ClientUpdate) {
	// create a new aclient instance
	pc, err := pclient.NewTCP(pcconf)
	if err != nil {
		log.Println(err)
		return
	}
	defer pc.Close()
	// success, so turn the update into json
	jupdate, err := json.Marshal(u)
	if err != nil {
		log.Println(err)
		return
	}
	req := []byte(cmd)
	req = append(req, 32)
	req = append(req, jupdate...)
	_, err = pc.Dispatch(req)
	if err != nil {
		log.Println("err dispatching", cmd, "for", svc, ":", err)
		return
	}
}
