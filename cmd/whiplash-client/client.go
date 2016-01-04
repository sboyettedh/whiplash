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
		{"ping": 11},
		{"ping": 13},
		{"ping": 17},
		{"ping": 19},
	}
)

func init() {
	flag.StringVar(&whipconf, "whipconf", "/etc/whiplash.conf", "Whiplash configuration file")
	hostname, _ = os.LookupEnv("HOSTNAME")
	nilPayload, _ := json.Marshal(nil)
}

func main() {
	flag.Parse()
	wl, err := whiplash.New(whipconf, true)
	if err != nil {
		log.Fatalf("error reading configuration file: %v\n", err)
	}
	sigchan := whiplash.AppSetup("whiplash-client", "0.2.0", aclient.Version)
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
	go pingSvcs(wl.Svcs, pingticker.C)

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

func pingSvcs(svcs map[string]*whiplash.Svc, tc <-chan time.Time) {
	for _ = range tc {
		for _, svc := range svcs {
			svc.Ping()
			if svc.Err != nil {
				log.Println("ping failed:", svc.Core.Name, svc.Err)
				continue
			}
			log.Println("sending ping request")
			sendData("ping", &whiplash.ClientRequest{Svc: svc.Core, Payload: nilPayload})
		}
	}
}

func sendData(cmd string, r *whiplash.ClientRequest) {
	ac, err := aclient.NewTCP(*acconf)
	if err != nil {
		log.Println(err)
		return
	}
	defer ac.Close()
	jreq, err := json.Marshal(r)
	if err != nil {
		log.Println(err)
		return
	}
	req := []byte(cmd)
	req = append(req, 32)
	req = append(req, jreq...)
	resp, err := ac.Dispatch(req)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(resp))
}
