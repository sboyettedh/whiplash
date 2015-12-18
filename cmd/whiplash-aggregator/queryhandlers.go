package main

import (
	"encoding/json"
	"log"

	"github.com/sboyettedh/whiplash"
)

func echoQHandler(args [][]byte) ([]byte, error) {
	resp := &whiplash.QueryResponse{
		Cmd: "echo",
		Subcmd: string(args[0]),
	}
	tmpdata, err := json.Marshal(nil) // TODO this goes away in real handlers
	resp.Data = tmpdata
	if err != nil {
		log.Println("echoQHandler problem", err)
	}
	for i := 1; i < len(args); i++ {
		resp.Args = append(resp.Args, string(args[i]))
	}
	return json.Marshal(resp)
}
