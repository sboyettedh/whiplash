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
	// current status of all OSDs
	osds map[string]*whiplash.Osd
	// the smaller map which we unmarshal JSON data about OSDs into
	josds map[string]*whiplash.Osd
	// map[cephstore][]osd - lets us do per-cephstore reporting easily
	osdmap map[string][]string
	// pre-rolled messages
	success = []byte("received")
)

func init() {
	flag.StringVar(&whipconf, "whipconf", "/etc/whiplash.json", "Whiplash configuration file")
	osds = make(map[string]*whiplash.Osd)
	josds = make(map[string]*whiplash.Osd)
	osdmap = make(map[string][]string)
}

func main() {
	// set up logfile
	f, err := os.Create("/var/log/whiplash.log")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	log.SetOutput(f)
	// write pidfile
	pidstr := strconv.Itoa(os.Getpid()) + "\n"
	err = ioutil.WriteFile("/var/run/whiplash-aggregator.pid", []byte(pidstr), 0644)
	if err != nil {
		log.Fatal(err)
	}
	// and register SIGINT/SIGTERM handler
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// parse flags and read the whiplash configuration
	flag.Parse()
	wl, err := whiplash.New(whipconf, false)
	if err != nil {
		log.Fatalf("could not read configuration file: %v\n", err)
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
	log.Println("Whiplash aggregator is listening.")

	// set up command handlers
	handlers := map[string]asock.DispatchFunc{
		"osdupdate": osdUpdate,
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

func osdUpdate(args [][]byte) ([]byte, error) {
	// unmarshal JSON data
	err := json.Unmarshal(args[0], &josds)
	if err != nil {
		return nil, err
	}
	// iterate over the vivified data
	//for osdname, osddata := range josds {
		// if we already know about the osd, update its data. if we don't, add it to 
	//	osd, ok := osds[osdname]
	//}
	return nil, nil
}
