package main

import (
	"encoding/json"
	"log"

	"github.com/sboyettedh/whiplash"
)

// These functions are the handlers for whiplash-aggregator's client
// side asock instance.

// pingHandler accepts and processes ping updates.
func pingHandler(args [][]byte) ([]byte, error) {
	upd := &whiplash.ClientUpdate{}
	err := json.Unmarshal(args[0], upd)
	if err != nil {
		return nil, err
	}
	// have we encountered this service before?
	ok := host2svcs.hostexists(upd.Svc.Host)
	if !ok {
		// no. add it to the host-to-svc map
		host2svcs.set(upd.Svc.Host, upd.Svc.Name)
	}
	// now add service to svcs and upds
	svcstat.set(upd.Svc.Name, upd.Svc)
	lastseen.set(upd.Svc.Name, "ping", upd.Time)
	log.Println("ping", upd.Svc.Name)
	return success, nil
}

// statHandler accepts and processes stat updates.
func statHandler(args [][]byte) ([]byte, error) {
/*	upd := &whiplash.ClientUpdate{}
	// unpack the update
	err := json.Unmarshal(args[0], upd)
	if err != nil {
		return nil, err
	}
	// unpack and update the status
	if upd.Svc.Type == whiplash.OSD {
		var stat *whiplash.OsdStat
		err = json.Unmarshal(upd.Payload, stat)
		osdstats[upd.Svc.Name] = stat
	}
	upds[upd.Svc.Name]["stat"] = upd.Time
*/
	return success, nil
}
