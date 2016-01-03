package main

import (
	"encoding/json"
	"log"

	"github.com/sboyettedh/whiplash"
)

// qhStubResponse creates and returns a stub QueryResponse for a
// successful request.
func qhStubResponse(cmd string, args [][]byte) *whiplash.QueryResponse {
	resp := &whiplash.QueryResponse{
		Code: 200,
		Cmd: cmd,
		Subcmd: string(args[0]),
	}
	for i := 1; i < len(args); i++ {
		resp.Args = append(resp.Args, string(args[i]))
	}
	return resp
}

// qhErrNoSubCmd should be called from a query handler when the subcmd
// (args[0]) is missing. It returns an appropriate error response.
func qhErrNoSubCmd(cmd string) ([]byte, error){
	errstr, _ := json.Marshal("no subcommand given")
	resp := &whiplash.QueryResponse{
		Code: 400,
		Cmd: cmd,
		Subcmd: "",
		Data: errstr,
	}
	return json.Marshal(resp)
}

// qhEcho is an example query handler. It simply creates an
// appropriate response and returns it.
func qhEcho(args [][]byte) ([]byte, error) {
	// we have to have a subcommand!
	if len(args) == 0 {
		return qhErrNoSubCmd("echo")
	}
	// gin up a blank success reponse
	resp := qhStubResponse("echo", args)
	// the Data field of a QueryResponse contains the result of the
	// handler's processing. it is also of type json.RawMessage, which
	// means that it's a byteslice of pre-marshalled data. here, we
	// catenate together the non-subcmd args and use that.
	respstr := ""
	for _, x := range resp.Args {
		respstr = respstr + string(x) + " "
	}
	// marshal the string we built, and make it resp.Data if there's
	// no problem.
	data, err := json.Marshal(respstr)
	if err != nil {
		log.Println("echoQHandler: ", err)
		return nil, err
	}
	resp.Data = data
	// json.Marshal returns ([]byte, error) which satisfies our
	// signature, so just return that!
	return json.Marshal(resp)
}
