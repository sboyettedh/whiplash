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

	"firepear.net/asock"
	"github.com/sboyettedh/whiplash"
)

var (
	// whiplash configuration file
	whipconf string
	// storage for current status of all reporting services
	svcs = map[string]*whiplash.SvcCore{}
	// map[cephstore][]svcname - lets us do per-cephstore reporting
	// until mon reporting of 'ceph osd tree' is in
	svcmap = map[string][]string{}
	// pre-rolled messages
	success = []byte("received")
)

func init() {
	flag.StringVar(&whipconf, "whipconf", "/etc/whiplash.json", "Whiplash configuration file")
}

func clientInit(fn string) (chan os.Signal, error) {
	// set up logfile
	f, err := os.Create("/var/log/" + fn + ".log")
	if err != nil {
		return nil, err
	}
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
	// parse flags and read the whiplash configuration
	flag.Parse()
	sigchan, err := clientInit("whiplash")
	if err != nil{
		log.Fatal(err)
	}
	log.Printf("whiplash-aggregator v%v beginning operations\n", whiplash.Version)
	log.Printf("asock version %s\n", asock.Version)

	wl, err := whiplash.New(whipconf, false)
	if err != nil {
		log.Fatalf("could not read configuration file: %v\n", err)
		os.Exit(1)
	}

	asconf := asock.Config{
		Sockname: wl.Aggregator.BindAddr + ":" + wl.Aggregator.BindPort,
		Msglvl: asock.All,
		Timeout: 100,
	}
	as, err := asock.NewTCP(asconf)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("listening for clients")

	// set up command handlers
	handlers := map[string]asock.DispatchFunc{
		"ping": pingHandler,
	}
	for name, handler := range handlers {
		err = as.AddHandler(name, "nosplit", handler)
		if err != nil {
			log.Fatal(err)
		}
	}

	// create a channel for the asock Msgr handler and launch it
	msgchan := make(chan error, 1)
	go msgHandler(as, msgchan)

	// the mainloop
	keepalive := true
	for keepalive {
		select {
		case msg := <-msgchan:
			// we've been handed a Msg over msgchan, which means that
			// our Asock has shut itself down for some reason. if this
			// were a more robust server, we would modularize Asock
			// creation and this eventloop, so that should we trap a
			// 599 we could spawn a new Asock and launch it in this
			// one's place. but we're just gonna exit this loop,
			// causing main() to terminate, and with it the server
			// instance.
			log.Println("Asock instance has shut down. Last Msg received was:")
			log.Println(msg)
			keepalive = false
			break
		case <- sigchan:
			// we've trapped a signal from the OS. tell our Asock to
			// shut down, but don't exit the eventloop because we want
			// to handle the Msgs which will be incoming.
			log.Println("OS signal received; shutting down")
			as.Quit()
		}
		// there's no default case in the select, as that would cause
		// it to be nonblocking. and that would cause main() to exit
		// immediately.
	}
}

func msgHandler(as *asock.Asock, msgchan chan error) {
	var msg *asock.Msg
	keepalive := true

	for keepalive {
		// wait on a Msg to arrive and do a switch based on status code
		msg = <-as.Msgr
		switch msg.Code {
		case 599:
			// 599 is "the Asock listener has died". this means we're
			// not accepting connections anymore. call as.Quit() to
			// clean things up, send the Msg to our main routine, then
			// kill this for loop
			as.Quit()
			keepalive = false
			msgchan <- msg
		case 199:
			// 199 is "we've been told to quit", so we want to break
			// out of the 'for' here as well
			keepalive = false
			msgchan <- msg
		default:
			// anything else we just log!
			log.Println(msg)
		}
	}
}

func pingHandler(args [][]byte) ([]byte, error) {
	req := &whiplash.Request{}
	json.Unmarshal(args[0], req)
	if svc, ok := svcs[req.Svc.Name]; !ok {
		log.Printf("adding svc %v", req.Svc.Name)
		// add sercice to svcs
		svcs[req.Svc.Name] = req.Svc
		// and to svcmap
		if _, ok := svcmap[req.Svc.Host]; !ok {
			log.Printf("adding host %v", req.Svc.Host)
			svcmap[req.Svc.Host] = []string{}
		}
		svcmap[req.Svc.Host] = append(svcmap[req.Svc.Host], req.Svc.Name)
	} else {
		// TODO handle updates
		log.Printf("got update from %v", svc.Name)
	}
	return []byte("ok"), nil
}
