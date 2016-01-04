package main

import (
	"encoding/json"
	"log"

	"github.com/sboyettedh/whiplash"
)

func pingHandler(args [][]byte) ([]byte, error) {
	req := &whiplash.ClientUpdate{}
	json.Unmarshal(args[0], req)
	if svc, ok := svcs[req.Svc.Name]; !ok {
		log.Println("adding svc", req.Svc.Name)
		// add service to svcs, upds
		svcs[req.Svc.Name] = req.Svc
		upds[req.Svc.Name] = map[string]int64{"ping": req.Time}
		// and to svcmap
		if _, ok := svcmap[req.Svc.Host]; !ok {
			log.Println("adding host", req.Svc.Host)
			svcmap[req.Svc.Host] = []string{}
		}
		svcmap[req.Svc.Host] = append(svcmap[req.Svc.Host], req.Svc.Name)
	} else {
		log.Println("updating", svc.Name)
		// TODO make version change an Event, once events are implemented
		svc.Version = req.Svc.Version
		svc.Reporting = req.Svc.Reporting
		upds[req.Svc.Name]["ping"] = req.Time
	}
	return []byte("ok"), nil
}
