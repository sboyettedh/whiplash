package main

import (
	"flag"
	"log"
	"sync"

	"firepear.net/asock"
	"github.com/sboyettedh/whiplash"
)

var (
	// whiplash configuration file
	whipconf = flag.String("config", "/etc/whiplash.conf", "Whiplash configuration file")

	// storage for current status of all reporting services
	svcstat = &svcStatus{m: make(map[string]*whiplash.SvcCore)}
	// host-to-service mapping
	host2svcs = &host2Svcs{m: make(map[string][]string)}
	// per-service update timestamps
	lastseen = &svcUpdates{m: make(map[string]map[string]int64)}

	// osd status info
	// NOT IN USE YET
	// osdstats = map[string]*whiplash.OsdStat{}

	// pre-rolled messages
	success = []byte("ok")
)

type svcStatus struct {
		sync.RWMutex
		m map[string]*whiplash.SvcCore
}
func (s *svcStatus) set(svcname string, stat *whiplash.SvcCore) {
	s.Lock()
	s.m[svcname] = stat
	s.Unlock()
}
func (s *svcStatus) get(svcname string) *whiplash.SvcCore {
	s.RLock()
	defer s.RUnlock()
	return s.m[svcname]
}

type host2Svcs struct {
	sync.RWMutex
	m map[string][]string
}
func (h2s *host2Svcs) set(hostname, svcname string) {
	h2s.Lock()
	_, ok := h2s.m[hostname]
	if !ok {
		h2s.m[hostname] = []string{}
	}
	h2s.m[hostname] = append(h2s.m[hostname], svcname)
	h2s.Unlock()
}
func (h2s *host2Svcs) getSvcs(hostname string) []string {
	h2s.RLock()
	defer h2s.RUnlock()
	return h2s.m[hostname]
}
func (h2s *host2Svcs) hostexists(hostname string) bool {
	h2s.RLock()
	defer h2s.RUnlock()
	_, ok := h2s.m[hostname]
	return ok
}

type svcUpdates struct {
	sync.RWMutex
	m map[string]map[string]int64
}
func (su *svcUpdates) set(svcname, handler string, time int64) {
	su.Lock()
	_, ok := su.m[svcname]
	if !ok {
		su.m[svcname] = map[string]int64{handler: time}
	} else {
		su.m[svcname][handler] = time
	}
	su.Unlock()
}
func (su *svcUpdates) get(svcname, handler string) int64 {
	su.RLock()
	defer su.RUnlock()
	return su.m[svcname][handler]
}
func (su *svcUpdates) getMostRecent(svcname string) (string, int64) {
	su.RLock()
	defer su.RUnlock()
	var handler string
	var time int64
	for h, t := range su.m[svcname] {
		if t > time {
			handler = h
			t = time
		}
	}
	return handler, time
}

func main() {
	// parse flags
	flag.Parse()
	// read the whiplash configuration
	wl, err := whiplash.New(*whipconf, false)
	if err != nil {
		log.Fatalf("error reading configuration file: %v\n", err)
	}
	// and do application initialization
	sigchan := whiplash.AppSetup("whiplash-aggregator", "0.1.1", asock.Pkgname, asock.Version)
	defer whiplash.AppCleanup("whiplash-aggregator")

	// setup the client asock instance. first set the msglvl, then
	// instantiate the asock.
	var msglvl int
	switch wl.Aggregator.MsgLvl {
	case "all":
		msglvl = asock.All
	case "conn":
		msglvl = asock.Conn
	case "error":
		msglvl = asock.Error
	case "fatal":
		msglvl = asock.Fatal
	}
	asconf := &asock.Config{
		Sockname: wl.Aggregator.BindAddr + ":" + wl.Aggregator.BindPort,
		Msglvl: msglvl,
		Timeout: wl.Aggregator.Timeout,
	}
	cas, err := asock.NewTCP(asconf)
	if err != nil {
		log.Fatal(err)
	}
	// and add command handlers to the asock instance
	handlers := map[string]asock.DispatchFunc{
		"ping": pingHandler,
		"stat": statHandler,
	}
	for name, handler := range handlers {
		err = cas.AddHandler(name, "nosplit", handler)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println("client asock instantiated")

	// now setup the query asock instance
	asconf = &asock.Config{
		Sockname: wl.Aggregator.BindAddr + ":" + wl.Aggregator.QueryPort,
		Msglvl: msglvl,
		Timeout: wl.Aggregator.QTimeout,
	}
	qas, err := asock.NewTCP(asconf)
	if err != nil {
		log.Fatal(err)
	}
	// add command handlers to the query asock instance
	handlers = map[string]asock.DispatchFunc{
		"echo": qhEcho,
	}
	for name, handler := range handlers {
		err = qas.AddHandler(name, "split", handler)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println("query asock instantiated")


	// create a channel for the client asock Msgr handler
	msgchan := make(chan error, 1)
	// and one for the query Msgr handler
	querychan := make(chan error, 1)
	// and launch them
	go msgHandler(cas, msgchan)
	go msgHandler(qas, querychan)
	log.Println("aggregator now listening")


	// this is the mainloop of the application.
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
		case msg := <-querychan:
			// the query handler has died. it should be safe to
			// restart.
			log.Println("Query asock instance has shut down. Last Msg received was:")
			log.Println(msg)
			log.Println("Restarting query asock...")
			// TODO what it says ^^there
		case <- sigchan:
			// we've trapped a signal from the OS. tell our Asock to
			// shut down, but don't exit the eventloop because we want
			// to handle the Msgs which will be incoming.
			log.Println("OS signal received; shutting down")
			cas.Quit()
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
