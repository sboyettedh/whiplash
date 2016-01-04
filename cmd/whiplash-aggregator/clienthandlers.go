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
		log.Printf("adding svc %v", req.Svc.Name)
		// add service to svcs
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
