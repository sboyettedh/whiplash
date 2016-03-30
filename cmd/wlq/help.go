package main

var helptext = map[string]string{
	"help": `The 'help' command prints this and other informational messages.`,
	"version": `The 'version' command prints the wlq and whiplash library versions.`,
	"main": `wlq is the Whiplash query tool.
Usage is: wlq <OPTS> [COMMAND] [SUBCOMMAND] <ARGS>

There are two global options:
    -c [FILE] Path to whiplash config file (default: /etc/whiplash.conf)
    -j        Output JSON data instead of a formatted report

To see available commands, run wlq with no arguments.

Do 'wlq help [COMMAND]' or 'wlq help [COMMAND] [SUBCOMMAND] for more detailed
information on usage of the commands and their subcommands.`,
	"crushreload": `The 'crushreload' command sends a request asking that the aggregator
reload the CRUSH map and refresh its cache of that data.`,
	"status": `The 'status' command fetches information on current cluster status from
the aggregator. By default this information is formatted and printed to the
terminal as a report. To get raw data, pass wlq the -j option.

To do anything useful, a subcommand must be specified. Available subcommands:
    cluster
    rack
    node
    osd
See 'wlq help status SUBCOMMAND' for information on a subcommand.`,
	"statuscluster":`The 'status cluster' command provides a look a the state of the cluster
as a whole.`,
	"statusrack":`Usage: wlq status rack [RACKNAME]
Shows an overview of the status of the given rack.`,
	"statusnode":`Usage: wlq status node [NODENAME]
Shows an overview of the status of the given cluster node.`,
	"statusosd":`Usage: wlq status osd [OSDID]
Shows detailed information about the given OSD.`,
}
