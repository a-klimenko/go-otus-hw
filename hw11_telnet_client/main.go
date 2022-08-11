package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var timeoutStr string

func init() {
	flag.StringVar(&timeoutStr, "timeout", "10s", "timeout to connect to the server (in seconds)")
}

func main() {
	var host, port string

	switch {
	case len(os.Args) < 3:
		log.Fatalf("not enough arguments")
	case len(os.Args) == 3:
		host = os.Args[1]
		port = os.Args[2]
	default:
		host = os.Args[2]
		port = os.Args[3]
	}

	address := net.JoinHostPort(host, port)
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		log.Fatalf("timeout is invalid: %s", err)
	}

	client := NewTelnetClient(
		address,
		timeout,
		os.Stdin,
		os.Stdout,
	)

	if err := client.Connect(); err != nil {
		log.Fatalf("Cannot connect: %v", err)
	}
	connectedMsg := []byte(fmt.Sprintf("...Connected to %s \n", address))
	if _, err := os.Stderr.Write(connectedMsg); err != nil {
		log.Fatal(err)
	}

	var closedMsg []byte
	defer func() {
		if err := client.Close(); err != nil {
			log.Fatalf("Cannot close connection: %v", err)
		}
		if _, err := os.Stderr.Write(closedMsg); err != nil {
			log.Fatalf("Cannot write to stderr: %v", err)
		}
	}()

	res := make(chan error)
	go func() {
		res <- client.Receive()
	}()

	go func() {
		res <- client.Send()
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT)
	defer signal.Stop(sigint)

	select {
	case <-sigint:
		closedMsg = []byte("...EOF \n")
	case <-res:
		closedMsg = []byte("...Connection was closed by peer \n")
		close(sigint)
	}
}
