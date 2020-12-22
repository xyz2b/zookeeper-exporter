package main

import (
	"log"
	"net"
	"time"
)

var (
	client net.Conn
)

func initClient() error {
	timeout := time.Duration(config.Timeout) * time.Second
	dialer := net.Dialer{Timeout: timeout}
	zkAddr, err := net.ResolveTCPAddr("tcp", config.ZkHost)
	if err != nil {
		log.Printf("warning: cannot resolve zk hostname '%s': %s", config.ZkHost, err)
		return err
	}

	conn, err := dialer.Dial("tcp", zkAddr.String())
	if err != nil {
		log.Printf("warning: cannot connect to %s: %v", config.ZkHost, err)
		return err
	}

	client = conn

	return nil
}
