package main

import (
	"fmt"
	"net"
	"time"
)

func askceph(sock, cmd string) ([]byte, error) {
	b1 := make([]byte, 64)
	var b2 []byte

	// make the connection
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, fmt.Errorf("could not connect to sock %v: %v\n", sock, err)
	}
	defer conn.Close()

	// send command to the admin socket
	conn.SetDeadline(time.Now().Add(250 * time.Millisecond))
	_, err = conn.Write([]byte(cmd + "\000"))
	if err != nil {
		return nil, fmt.Errorf("could not write to %v: %v\n", sock, err)
	}

	// now read what we got back.
	conn.SetDeadline(time.Now().Add(250 * time.Millisecond))
	for {
		n, err := conn.Read(b1)
		if err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("could not read from %v: %v\n", sock, err)
		}
		// since the admin-daemon closes the connection as soon as
		// it's done writing, there's no EOM to watch for. you just
		// read until there's nothing left, and then yo're done.
		if n == 0 {
			break
		}
		b2 = append(b2, b1[:n]...)
	}
	return b2, err
}
