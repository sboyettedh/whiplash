package whiplash

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	DEBUG = iota
	ERROR
)

// AppSetup switches logging to a file (based on application name) and
// does pidfile handling.
func AppSetup(appname, appver, asockver string) (chan os.Signal) {
	// First, see if an existing pidfile can be read
	pidfile := "/var/run/" + appname + ".pid"
	pb, err := ioutil.ReadFile(pidfile)
	if err == nil {
		// the only safe thing to do, given an extant pidfile, is
		// refuse to start.
		log.Fatalf("pidfile found at %s; pid %s; refusing to start",
			pidfile, string(pb[:len(pb) - 1]))
	}

	// No pidfile, or stale pidfile. Carry on: open logfile in append
	// mode, or create if it doesn't exist
	f, err := os.OpenFile("/var/log/" + appname + ".log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// set logging to the logfile
	log.SetOutput(f)
	// write startup messages
	log.Printf("====================================== %s %s beginning operations\n", appname, appver)
	log.Printf("asock/client v%s; whiplash lib v%s\n", asockver, Version)

	// write new pidfile
	pidstr := strconv.Itoa(os.Getpid()) + "\n"
	err = ioutil.WriteFile(pidfile, []byte(pidstr), 0644)
	if err != nil {
		log.Fatal(err)
	}
	// and register SIGINT/SIGTERM handler
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	return sigchan
}

// AppCleanup deletes an app's pidfile
func AppCleanup(appname string) {
	pidfile := "/var/run/" + appname + ".pid"
	err := os.Remove(pidfile)
	if err != nil {
		log.Printf("couldn't remove %s: %s", pidfile, err)
	}
}
