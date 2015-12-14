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

func AppSetup(appname, appver, asockver string) (chan os.Signal) {
	// logfile handling
	f, err := os.OpenFile("/var/log/" + appname + ".log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	log.Printf("============================== %s %s beginning operations\n", appname, appver)
	log.Printf("asock version %s; whiplash version\n", asockver, Version)


	// pidfile handling
	pidstr := strconv.Itoa(os.Getpid()) + "\n"
	err = ioutil.WriteFile("/var/run/" + appname + ".pid", []byte(pidstr), 0644)
	if err != nil {
		log.Fatal(err)
	}
	// and register SIGINT/SIGTERM handler
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	return sigchan
}
