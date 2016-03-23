package main

var helptext = map[string]string{
	"help": `The 'help' command prints this and other informational messages.`,
	"version": `The 'version' command prints the wlq and whiplash library versions.`,
	"main": `wlq is the Whiplash query tool.
Usage is: wlq <OPTS> [COMMAND] [SUBCOMMAND] <ARGS>

There are two global options:
    -c [FILE] Path to whiplash config file (default: /etc/whiplash.conf)
    -j        Output JSON data instead of a formatted report

Available commands are:
    help
    version
    status

Do 'wlq help [COMMAND]' or 'wlq help [COMMAND] [SUBCOMMAND] for more detailed
information on usage of the commands and their subcommands.`,
	"status": `The 'status' command fetches information on current cluster status from
the aggregator. By default this information is formatted and printed to the
terminal as a report. To get raw data, pass wlq the -j option.

To do anything useful, a subcommand must be specified. Available subcommands:
    cluster
    rack
    node
    osd
See 'wlq help status SUBCOMMAND' for information on a subcommand.`,
	"statuscluster":``,
	"statusrack":``,
	"statusnode":``,
	"statusosd":``,
}
