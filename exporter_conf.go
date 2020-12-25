package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func init() {
	//RegisterExporter("overview", newExporterOverview)
}

var (
	confGaugeVec = map[string]*prometheus.GaugeVec{
		"maxClientCnxns":            newGaugeVec("max_connections", "max of connections.", "node"),
	}
)

type exporterConf struct {
	confGauge 	map[string]*prometheus.GaugeVec
	nodeInfo        NodeInfo
}

//NodeInfo presents the name and version of fetched rabbitmq
type NodeInfo struct {
	Node            string
}

func newExporterConf() *exporterConf {
	confaugeVecActual := confGaugeVec

	return &exporterConf{
		confGauge: confaugeVecActual,
		nodeInfo:        NodeInfo{},
	}
}

func (e exporterConf) NodeInfo() NodeInfo {
	return e.nodeInfo
}

func makeConfStatsInfo(body []byte, labels ...string) []StatsInfo {
	var q []StatsInfo

	statsinfo := StatsInfo{}
	statsinfo.metrics = make(MetricMap)

	reply := string(body)

	lines := strings.Split(reply, "\n")

	for _, line := range lines {
		if strings.Contains(line, "membership:") || line == "" {
			continue
		}

		kv := strings.Split(line, "=")

		statsinfo.metrics[kv[0]] = kv[1]
	}

	q = append(q, statsinfo)

	return q
}

func (e *exporterConf) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	zkConfData, err := getStatsInfo(makeConfStatsInfo, "conf")
	if err != nil {
		return err
	}

	port := zkConfData[0].metrics["clientPort"]

	serverId := zkConfData[0].metrics["serverId"]

	host := ""
	if zkHost, ok := zkConfData[0].metrics["server." + serverId]; ok {
		host = strings.Split(zkHost, ":")[0] + port
	} else {
		host = config.ZkHost
	}

	e.nodeInfo.Node = host

	log.WithField("confData", zkConfData).Debug("Conf data")

	for key, gauge := range e.confGauge {
		for _, conf := range zkConfData {
			if value, ok := conf.metrics[key]; ok {
				log.WithFields(log.Fields{"key": key, "value": value}).Debug("Set conf metric for key")
				v, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.WithFields(log.Fields{"key": key, "value": value}).Error("conv value to float64 failed")
					continue
				}
				gaugeVecWithLabelValues(gauge, e.NodeInfo().Node).Set(v)
			}
		}
	}

	if ch != nil {
		for _, gauge := range e.confGauge {
			gauge.Collect(ch)
		}
	}
	return nil
}

func (e exporterConf) Describe(ch chan<- *prometheus.Desc) {
	for _, gauge := range e.confGauge {
		gauge.Describe(ch)
	}

}