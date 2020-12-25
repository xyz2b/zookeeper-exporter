package main

import (
	"io/ioutil"
	"log"
	"net"
	"time"
)

func newClient() (net.Conn, error) {
	timeout := time.Duration(config.Timeout) * time.Second
	dialer := net.Dialer{Timeout: timeout}
	zkAddr, err := net.ResolveTCPAddr("tcp", config.ZkHost)
	if err != nil {
		log.Printf("warning: cannot resolve zk hostname '%s': %s", config.ZkHost, err)
		return nil, err
	}

	conn, err := dialer.Dial("tcp", zkAddr.String())
	if err != nil {
		log.Printf("warning: cannot connect to %s: %v", config.ZkHost, err)
		return nil, err
	}

	return conn, nil
}

func getStatsInfo(makeStatsInfo func(body []byte, labels ...string) []StatsInfo, apiEndpoint string, labels ...string) ([]StatsInfo, error) {
	var q []StatsInfo

	client, err := newClient()
	defer client.Close()
	if err != nil {
		return nil, err
	}
	reply, err := sendZookeeperCmd(client, config.ZkHost, apiEndpoint)
	if err != nil {
		return q, err
	}

	q = makeStatsInfo(reply, labels...)

	return q, nil
}

func sendZookeeperCmd(conn net.Conn, host, cmd string) ([]byte, error) {
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		log.Printf("warning: failed to send '%s' to '%s': %s", cmd, host, err)
		return nil, err
	}

	res, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Printf("warning: failed read '%s' response from '%s': %s", cmd, host, err)
		return nil, err
	}

	return res, nil
}