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
	svcs.RLock()
	_, ok := svcs.m[upd.Svc.Name]
	svcs.RUnlock()
	if !ok {
		// no. add it (and its host if needed) to the svcmap
		svcmap.RLock()
		_, ok := svcmap.m[upd.Svc.Host]
		svcmap.RUnlock()
		if !ok {
			// host not found; add host and service
			svcmap.Lock()
			svcmap.m[upd.Svc.Host] = []string{}
			svcmap.m[upd.Svc.Host] = append(svcmap.m[upd.Svc.Host], upd.Svc.Name)
			svcmap.Unlock()
			log.Println("added host", upd.Svc.Host)
		} else {
			// host found; just add the service (duplicating the
			// append prevents back-to-back locking)
			svcmap.Lock()
			svcmap.m[upd.Svc.Host] = append(svcmap.m[upd.Svc.Host], upd.Svc.Name)
			svcmap.Unlock()
		}
		// now add service to svcs and upds
		svcs.Lock()
		svcs.m[upd.Svc.Name] = upd.Svc
		svcs.Unlock()
		upds.Lock()
		upds.m[upd.Svc.Name] = map[string]int64{"ping": upd.Time}
		upds.Unlock()
		log.Println("added svc", upd.Svc.Name)
	} else {
		// we have seen the service before. update it!
		// TODO make version change an Event, once events are implemented
		svcs.Lock()
		svcs.m[upd.Svc.Name].Version = upd.Svc.Version
		svcs.m[upd.Svc.Name].Reporting = upd.Svc.Reporting
		svcs.Unlock()
		upds.Lock()
		upds.m[upd.Svc.Name]["ping"] = upd.Time
		upds.Unlock()
		log.Println("updated svc", upd.Svc.Name)
	}
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
