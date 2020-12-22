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

// TODO: zk四字命令获取指标信息，先梳理有哪些四字命令，能够获取什么指标，然后选出需要的指标
func getStatsInfo()  {
	
}