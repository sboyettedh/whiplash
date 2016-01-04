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
	if svc, ok := svcs[upd.Svc.Name]; !ok {
		log.Println("adding svc", upd.Svc.Name)
		// add service to svcs, upds
		svcs[upd.Svc.Name] = upd.Svc
		upds[upd.Svc.Name] = map[string]int64{"ping": upd.Time}
		// and to svcmap
		if _, ok := svcmap[upd.Svc.Host]; !ok {
			log.Println("adding host", upd.Svc.Host)
			svcmap[upd.Svc.Host] = []string{}
		}
		svcmap[upd.Svc.Host] = append(svcmap[upd.Svc.Host], upd.Svc.Name)
	} else {
		log.Println("updating", svc.Name)
		// TODO make version change an Event, once events are implemented
		svc.Version = upd.Svc.Version
		svc.Reporting = upd.Svc.Reporting
		upds[upd.Svc.Name]["ping"] = upd.Time
	}
	return success, nil
}

// statHandler accepts and processes stat updates.
func statHandler(args [][]byte) ([]byte, error) {
	upd := &whiplash.ClientUpdate{}
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

	return success, nil
}
