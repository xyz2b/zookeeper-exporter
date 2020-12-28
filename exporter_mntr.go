package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

func init() {
	RegisterExporter("mntr", newExporterMntr)
}

type exporterMntr struct {
	mntrGauge		map[string]*prometheus.GaugeVec
}

func newExporterMntr() Exporter {
	mntrGaugeVecActual := map[string]*prometheus.GaugeVec{
		"zk_num_alive_connections":            newGaugeVec("connections", "the number of connections.", "node"),
		"zk_server_state":					   newGaugeVec("server_is_leader", "server mode(follower/leader).", "node"),
		"zk_min_latency":					   newGaugeVec("min_latency", "Minimum Latency.", "node"),
		"zk_avg_latency":				       newGaugeVec("avg_latency", "Average Latency.", "node"),
		"zk_max_latency":                      newGaugeVec("max_latency", "Maximum Latency.", "node"),
		"zk_open_file_descriptor_count":	   newGaugeVec("open_file_descriptor_count", "Number of open file descriptors.", "node"),
		"zk_max_file_descriptor_count":		   newGaugeVec("max_file_descriptor_count", "Maximum number of file descriptors.", "node"),
		"zk_outstanding_requests":			   newGaugeVec("outstanding_requests", "Stacked requests.", "node"),
		"zk_approximate_data_size":			   newGaugeVec("approximate_data_size", "Data size.", "node"),
		"zk_packets_sent":					   newGaugeVec("packets_sent", "Number of packets sent.", "node"),
		"zk_packets_received": 				   newGaugeVec("packets_received", "Number of packets received.", "node"),
		"zk_followers":						   newGaugeVec("followers", "Number of follower(Only leader have).", "node"),
		"zk_synced_followers":				   newGaugeVec("synced_followers", "Number of synchronized follower(Only leader have).", "node"),
		"zk_pending_syncs":					   newGaugeVec("pending_syncs", "Number of ready to sync.", "node"),
		"zk_last_proposal_size":			   newGaugeVec("last_proposal_size", "The size of the last Proposal message.", "node"),
		"zk_max_proposal_size":				   newGaugeVec("max_proposal_size", "The size of the maximum Proposal message.", "node"),
		"zk_min_proposal_size":				   newGaugeVec("min_proposal_size", "The size of the minimum Proposal message.", "node"),
		"zk_cnt_node_changed_watch_count":     newGaugeVec("cnt_node_changed_watch_count", "the changed watch count", "node"),
	}

	return &exporterMntr{
		mntrGauge: mntrGaugeVecActual,
	}
}

func makeMntrStatsInfo(body []byte, labels ...string) []StatsInfo {
	var q []StatsInfo

	statsinfo := StatsInfo{}
	statsinfo.metrics = make(MetricMap)

	reply := string(body)
	lines := strings.Split(reply, "\n")

	for _, line := range lines {
		if line ==  "" {
			continue
		}

		line = strings.Replace(line, "\t", " ", -1)
		kv := strings.Split(line, " ")

		if kv[0] == "zk_server_state" {
			if kv[1] == "leader" {
				kv[1] = "1"
			} else {
				kv[1] = "0"
			}
		}

		statsinfo.metrics[kv[0]] = kv[1]
	}

	q = append(q, statsinfo)

	return q
}

func (e exporterMntr) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	node := ""
	if n, ok := ctx.Value(nodeName).(string); ok {
		node = n
	}

	zkMntrData, err := getStatsInfo(makeMntrStatsInfo, "mntr")
	if err != nil {
		return err
	}

	log.WithField("mntrData", zkMntrData).Debug("mntr data")

	for key, gauge := range e.mntrGauge {
		for _, mntr := range zkMntrData {
			if value, ok := mntr.metrics[key]; ok {
				log.WithFields(log.Fields{"key": key, "value": value}).Debug("Set mntr metric for key")
				v, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.WithFields(log.Fields{"key": key, "value": value}).Error("conv value to float64 failed")
					continue
				}
				gaugeVecWithLabelValues(gauge, node).Set(v)
			}
		}
	}

	if ch != nil {
		for _, gauge := range e.mntrGauge {
			gauge.Collect(ch)
		}
	}
	return nil
}

func (e exporterMntr) Describe(ch chan<- *prometheus.Desc) {
	for _, gauge := range e.mntrGauge {
		gauge.Describe(ch)
	}

}