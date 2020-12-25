package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"strconv"
)

func init() {
	RegisterExporter("ruok", newExporterRuok)
}

var (
	ruokGaugeVec = map[string]*prometheus.GaugeVec{
		"ruok":            newGaugeVec("up", "the status of zookeeper service.", "node"),
	}
)

type exporterRuok struct {
	ruokGauge		map[string]*prometheus.GaugeVec
}

func newExporterRuok() Exporter {
	ruokaugeVecActual := ruokGaugeVec

	return &exporterRuok{
		ruokGauge: ruokaugeVecActual,
	}
}

func makeRuokStatsInfo(body []byte, labels ...string) []StatsInfo {
	var q []StatsInfo

	statsinfo := StatsInfo{}
	statsinfo.metrics = make(MetricMap)

	reply := string(body)

	up := "0"
	if reply == "imok" {
		up = "1"
	}

	statsinfo.metrics["ruok"] = up

	q = append(q, statsinfo)

	return q
}

func (e exporterRuok) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	node := ""
	if n, ok := ctx.Value(nodeName).(string); ok {
		node = n
	}

	zkRuokData, err := getStatsInfo(makeRuokStatsInfo, "ruok")
	if err != nil {
		return err
	}

	log.WithField("ruokData", zkRuokData).Debug("ruok data")

	for key, gauge := range e.ruokGauge {
		for _, ruok := range zkRuokData {
			if value, ok := ruok.metrics[key]; ok {
				log.WithFields(log.Fields{"key": key, "value": value}).Debug("Set ruok metric for key")
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
		for _, gauge := range e.ruokGauge {
			gauge.Collect(ch)
		}
	}
	return nil
}

func (e exporterRuok) Describe(ch chan<- *prometheus.Desc) {
	for _, gauge := range e.ruokGauge {
		gauge.Describe(ch)
	}

}