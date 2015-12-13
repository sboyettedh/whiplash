package whiplash

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func AppSetup(fn string) (chan os.Signal, error) {
	// set up logfile
	f, err := os.Create("/var/log/" + fn + ".log")
	if err != nil {
		return nil, err
	}
	log.SetOutput(f)
	// write pidfile
	pidstr := strconv.Itoa(os.Getpid()) + "\n"
	err = ioutil.WriteFile("/var/run/" + fn + ".pid", []byte(pidstr), 0644)
	if err != nil {
		return nil, err
	}
	// and register SIGINT/SIGTERM handler
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	return sigchan, err
}
