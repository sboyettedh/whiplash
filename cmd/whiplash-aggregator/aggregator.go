package main

import (
	//"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"firepear.net/asock"
	"github.com/sboyettedh/whiplash"
)

var (
	whipconf string
	osds map[string]*whiplash.Osd // current status of all OSDs
	osdmap map[string][]string // OSDs in each cephstore
)

func init() {
	flag.StringVar(&whipconf, "whipconf", "/etc/whiplash.json", "Whiplash configuration file")
}

func main() {
	flag.Parse()
	wl, err := whiplash.New(whipconf, false)
	if err != nil {
		log.Printf("%v: could not read configuration file: %v\n", os.Args[0], err)
		os.Exit(1)
	}

	// let's give ourselves a way to shut down. we'll listen for
	// SIGINT and SIGTERM, so we can behave like a proper service
	// (mostly -- we're not writing out a pidfile). anyway, to do that
	// we need a channel to recieve signal notifications on.
	sigchan := make(chan os.Signal, 1)
	// and then we register sigchan to listen for the signals we want.
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	// we now respond properly to 'kill' calls to our pid, and to C-c
	// at the terminal we're running in.

	asconf := asock.Config{
		Sockname: wl.Aggregator.BindAddr + ":" + wl.Aggregator.BindPort,
		Msglvl: asock.All,
		Timeout: 100,
	}
	as, err := asock.NewTCP(asconf)
	if err != nil {
		panic(err)
	}
	log.Println("Whiplash aggregator is listening.")

	// set up command handlers
	handlers := map[string]asock.DispatchFunc{
		"osdupdate": osdUpdate,
	}
	for name, handler := range handlers {
		err = as.AddHandler(name, "nosplit", handler)
		if err != nil {
			panic(err)
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
	log.Println("Got payload")
	return nil, nil
}
