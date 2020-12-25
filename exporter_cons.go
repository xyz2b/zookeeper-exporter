package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

func init() {
	RegisterExporter("cons", newExporterCons)
}

var (
	consGaugeVec = map[string]*prometheus.GaugeVec{
		"queued":            newGaugeVec("client_queued", "Client queue.", "node", "client"),
		"recved":            newGaugeVec("client_recved", "Number of packets received by the client.", "node", "client"),
		"sent":            newGaugeVec("client_sent", "Number of packets sent by the client.", "node", "client"),
		"sid":            newGaugeVec("client_sid", "Client Session Id.", "node", "client"),
		"lop":            newGaugeVec("client_lop", "Client last operation instructions.", "node", "client"),
		"est":            newGaugeVec("client_est", "Client connection timestamp.", "node", "client"),
		"to":            newGaugeVec("client_to", "Client connection timeout.", "node", "client"),
		"lcxid":            newGaugeVec("client_lcxid", "The last id of the client (no specific id confirmed).", "node", "client"),
		"lzxid":            newGaugeVec("client_lzxid", "The last id of the client (state change id).", "node", "client"),
		"lresp":            newGaugeVec("client_lresp", "Client last response timestamp.", "node", "client"),
		"llat":				newGaugeVec("client_llat", "Client last delay.", "node", "client"),
		"minlat":            newGaugeVec("client_minlat", "Client Minimum delay.", "node", "client"),
		"avglat":            newGaugeVec("client_avglat", "Client Average delay.", "node", "client"),
		"maxlat":            newGaugeVec("client_maxlat", "Client Maximum delay.", "node", "client"),
	}
)

type exporterCons struct {
	consGauge		map[string]*prometheus.GaugeVec
}

func newExporterCons() Exporter {
	consaugeVecActual := consGaugeVec

	return &exporterCons{
		consGauge: consaugeVecActual,
	}
}

func makeConsStatsInfo(body []byte, labels ...string) []StatsInfo {
	var q []StatsInfo

	statsinfo := StatsInfo{}
	statsinfo.labels = make(map[string]string)
	statsinfo.metrics = make(MetricMap)

	reply := string(body)

	lines := strings.Split(reply, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		cons := strings.Split(reply, "/")[1]
		s := strings.Split(cons, "(")
		client := s[0]

		statsinfo.labels["client"] = client

		m := strings.Split(s[1], ")")[0]

		metrics := strings.Split(m, ",")

		for _, metric := range metrics {
			kv := strings.Split(metric, "=")
			statsinfo.metrics[kv[0]] = kv[1]
		}
	}

	q = append(q, statsinfo)

	return q
}

func (e exporterCons) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	node := ""
	if n, ok := ctx.Value(nodeName).(string); ok {
		node = n
	}

	zkConsData, err := getStatsInfo(makeConsStatsInfo, "cons")
	if err != nil {
		return err
	}

	log.WithField("consData", zkConsData).Debug("cons data")

	for key, gauge := range e.consGauge {
		for _, cons := range zkConsData {
			client := cons.labels["client"]
			if value, ok := cons.metrics[key]; ok {
				log.WithFields(log.Fields{"key": key, "value": value}).Debug("Set ruok metric for key")
				v, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.WithFields(log.Fields{"key": key, "value": value}).Error("conv value to float64 failed")
					continue
				}
				gaugeVecWithLabelValues(gauge, node, client).Set(v)
			}
		}
	}

	if ch != nil {
		for _, gauge := range e.consGauge {
			gauge.Collect(ch)
		}
	}
	return nil
}

func (e exporterCons) Describe(ch chan<- *prometheus.Desc) {
	for _, gauge := range e.consGauge {
		gauge.Describe(ch)
	}

}