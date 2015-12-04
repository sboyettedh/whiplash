package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

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

func startup(fn string) (chan os.Signal, error) {
	// set up logfile
	f, err := os.Create("/var/log/" + fn + ".log")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.SetOutput(f)
	// write pidfile
	pidstr := strconv.Itoa(os.Getpid()) + "\n"
	err = ioutil.WriteFile("/var/run/" + fn + ".pid", []byte(pidstr), 0644)
	if err != nil {
		return nil, err
	}
	// and register SIGINT/SIGTERM handler
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	return sigchan, err
}

func main() {
	sigchan, err := startup("whiplash-client")
	if err != nil{
		log.Fatal(err)
	}

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
		log.Println(err)
	}
	defer ac.Close()
	josd, err := json.Marshal(osds)
	if err != nil {
		log.Println(err)
	}
	req := []byte("osdupdate ")
	req = append(req, josd...)
	resp, err := ac.Dispatch(req)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(resp))

	// the mainloop
	keepalive := true
	for keepalive {
		select {
		case <- sigchan:
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

