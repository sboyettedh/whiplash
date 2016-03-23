package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"firepear.net/aclient"
	"firepear.net/gaot"
	"github.com/sboyettedh/whiplash"
)

var (
	whipconf string
	dumpjson bool
	commands *gaot.Node
	args []string
)

func init() {
	// setup options vars
	flag.StringVar(&whipconf, "config", "/etc/whiplash.conf", "Whiplash configuration file")
	flag.BoolVar(&dumpjson, "j", false, "Output query response as raw JSON")
	// load up command structure into trie
	commands = gaot.NewFromString("status", nil)
	commands.InsertString("version", nil)
	cmdtail := commands.FindString("status")
	cmdtail.InsertString("cluster", nil)
	cmdtail.InsertString("rack", nil)
	cmdtail.InsertString("node", nil)
	cmdtail.InsertString("osd", nil)
}

func main() {
	flag.Parse()
	args = flag.Args()
	// read whiplash config. genconf should be FALSE in the
	// whiplash.New call for a wlq instance: we don't expect to hacve
	// ceph services around on a machine running a query.
	wl, err := whiplash.New(whipconf, false)
	if err != nil {
		quit(fmt.Errorf("can't read configuration file: %s\n", err))
	}

	// validate user input
	err = validateInput()
	if err != nil {
		quit(err)
	}
	// handle non-networked commands
	switch args[0] {
	case "version":
		showVersion()
		quit(nil)
	}


	// set up configuration and create aclient instance
	acconf := &aclient.Config{
		Addr: wl.Aggregator.BindAddr + ":" + wl.Aggregator.QueryPort,
		Timeout: 100,
	}
	c, err := aclient.NewTCP(acconf)
	if err != nil {
		quit(fmt.Errorf("can't connect to aggregator: %s", err))
	}
	defer c.Close()

	// stitch together the non-option arguments into our request
	req := strings.Join(flag.Args(), " ")
	// and dispatch it to the server!
	respj, err := c.Dispatch([]byte(req))
	if err != nil {
		quit(fmt.Errorf("sending request to aggregator failed: %s", err))
	}

	// vivify response and handle errors
	resp := new(whiplash.QueryResponse)
	err = json.Unmarshal(respj, &resp)
	if err != nil {
		quit(fmt.Errorf("failure unmarshaling json\nerror: %s\njson: %s", err, string(respj)))
	}
	if resp.Code >= 400 {
		quit(fmt.Errorf("there was a problem with the request:\n%s", string(respj)))
	}

	// if -j has been specified, print the raw response data and exit
	if dumpjson {
		fmt.Println(string(resp.Data))
		os.Exit(0)
	}

	// else, we have to hand off to a pretty-printing routine
	// TODO write a pretty-printing routine
}

func showVersion() {
	fmt.Printf("wlq v0.1.0 (whiplash v%s)\n", whiplash.Version)
	quit(nil)
}

func validateInput() error {
	// get list of top-level completions (commands)
	cmdlist := commands.FirstCompletions()
	cmds := ""
	for _, known := range cmdlist {
		cmds = cmds + known.Word + " "
	}
	// get our arguments
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("no command given\n\tknown commands: %s", cmds)
	}
	// see if we know the command
	cmd := commands.FindString(args[0])
	if cmd == nil {
		// no. list known commands
		return fmt.Errorf("unknown command: '%s'\n\tknown commands: %s", args[0], cmds)
	} else if cmd.Word == "" {
		// partial match: show word completions from here
		cmdlist = cmd.FirstCompletions()
		maybe := ""
		for _, known := range cmdlist {
			maybe = maybe + known.Word + " "
		}
		return fmt.Errorf("unknown command %s\n\tdid you mean? %s", args[0], maybe)
	}
	// yes. check the subcommand
	cmdlist = cmd.Completions()
	if len(cmdlist) == 0 {
		return nil
	}
	subcmds := ""
	for _, known := range cmdlist {
		subcmds = subcmds + known.Word + " "
	}
	if len(args) == 1 {
		return fmt.Errorf("no subcommand given\n\tsubcommands for %s: %s", args[0], subcmds)
	}
	subcmd := cmd.FindString(args[1])
	if subcmd == nil {
		return fmt.Errorf("unknown subcommand: '%s'\n\tsubcommands for %s: %s",
			args[1], args[0], subcmds)
	}
	return nil
}

func quit(err error) {
	if err != nil {
		fmt.Printf("wlq: %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
