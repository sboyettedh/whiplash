package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"os"
	"time"

	"firepear.net/aclient"
	"github.com/sboyettedh/whiplash"
)

var (
	whipconf string
	hostname string
	acconf *aclient.Config
	// nil payload for pings
	nilPayload json.RawMessage
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

func init() {
	flag.StringVar(&whipconf, "whipconf", "/etc/whiplash.conf", "Whiplash configuration file")
	hostname, _ = os.LookupEnv("HOSTNAME")
	nilPayload, _ = json.Marshal(nil)
}

func main() {
	flag.Parse()
	wl, err := whiplash.New(whipconf, true)
	if err != nil {
		log.Fatalf("error reading configuration file: %v\n", err)
	}
	sigchan := whiplash.AppSetup("whiplash-client", "0.3.0", aclient.Version)
	defer whiplash.AppCleanup("whiplash")


	// need an aclient configuration to talk to the aggregator with
	acconf = &aclient.Config{
		Addr: wl.Aggregator.BindAddr + ":" + wl.Aggregator.BindPort,
		Timeout: 100,
	}

	// decide what notification interval to use
	rand.Seed(time.Now().UnixNano())
	intv = rand.Intn(len(intvs))
	log.Printf("using interval set: %q\n", intvs[intv])
	// create tickers and launch monitor funcs
	pingticker := time.NewTicker(time.Second * intvs[intv]["ping"])
	statticker := time.NewTicker(time.Second * intvs[intv]["stat"])
	go pingSvcs(wl.Svcs, pingticker.C)
	go pingSvcs(wl.Svcs, statticker.C)

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
		for _, svc := range svcs {
			// ping, then send data if there are no issues
			svc.Ping()
			if svc.Err != nil {
				log.Println("ping failed:", svc.Core.Name, svc.Err)
				continue
			}
			sendData("ping", &whiplash.ClientUpdate{
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
		for _, svc := range svcs {
			statdata = svc.Stat()
			// TODO handle MONs and RGWs
			if svc.Err != nil {
				log.Println("ping failed:", svc.Core.Name, svc.Err)
				continue
			}
			sendData("stat", &whiplash.ClientUpdate{
				Time: time.Now().Unix(),
				Svc: svc.Core,
				Payload: statdata,
			})
		}
	}
}

// sendData handles the actual sending of data to the aggregator.
func sendData(cmd string, u *whiplash.ClientUpdate) {
	// create a new aclient instance
	ac, err := aclient.NewTCP(*acconf)
	if err != nil {
		log.Println(err)
		return
	}
	defer ac.Close()
	// success, so turn the update into json
	jupdate, err := json.Marshal(u)
	if err != nil {
		log.Println(err)
		return
	}
	req := []byte(cmd)
	req = append(req, 32)
	req = append(req, jupdate...)
	log.Println("sending update:", cmd)
	resp, err := ac.Dispatch(req)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("got response:", string(resp))
}
