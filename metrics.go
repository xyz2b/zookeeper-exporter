package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "zookeeper"
)

var (
	extraLabelNames []string
	extraLabelValues []string
)

type MetricMap map[string]string

type StatsInfo struct {
	labels  map[string]string
	metrics MetricMap
}

func initExtraLabels() {
	if config.ExtraLabels != nil {
		for _, extraLabel := range config.ExtraLabels {
			for k, v := range extraLabel {
				extraLabelNames = append(extraLabelNames, k)
				extraLabelValues = append(extraLabelValues, v)
			}
		}
	}
}

func newGaugeVec(metricName string, docString string, labelNames ...string) *prometheus.GaugeVec {
	if labelNames != nil {
		labelNames = append(labelNames, extraLabelNames...)
	} else {
		labelNames = extraLabelNames
	}

	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metricName,
			Help:      docString,
		},
		labelNames,
	)
}

func newCounterVec(metricName string, docString string, labelNames ...string) *prometheus.CounterVec {
	if labelNames != nil {
		labelNames = append(labelNames, extraLabelNames...)
	} else {
		labelNames = extraLabelNames
	}

	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      metricName,
			Help:      docString,
		},
		labelNames,
	)
}

func newDesc(metricName string, docString string, labelNames ...string) *prometheus.Desc {
	if labelNames != nil {
		labelNames = append(labelNames, extraLabelNames...)
	} else {
		labelNames = extraLabelNames
	}

	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", metricName),
		docString,
		labelNames,
		nil)
}


func counterVecWithLabelValues(v *prometheus.CounterVec, labelValues ...string) prometheus.Counter {
	if labelValues != nil {
		labelValues = append(labelValues, extraLabelValues...)
	} else {
		labelValues = extraLabelValues
	}

	counter := v.WithLabelValues(labelValues...)
	return counter
}

func gaugeVecWithLabelValues(v *prometheus.GaugeVec, labelValues ...string) prometheus.Gauge {
	if labelValues != nil {
		labelValues = append(labelValues, extraLabelValues...)
	} else {
		labelValues = extraLabelValues
	}

	counter := v.WithLabelValues(labelValues...)
	return counter
}

func mustNewConstHistogram(
	desc *prometheus.Desc,
	count uint64,
	sum float64,
	buckets map[float64]uint64,
	labelValues ...string,
) prometheus.Metric {
	if labelValues != nil {
		labelValues = append(labelValues, extraLabelValues...)
	} else {
		labelValues = extraLabelValues
	}

	metric := prometheus.MustNewConstHistogram(
		desc, count, sum, buckets, labelValues...
	)
	return metric
}

func mustNewConstMetric(desc *prometheus.Desc, valueType prometheus.ValueType, value float64, labelValues ...string) prometheus.Metric {
	if labelValues != nil {
		labelValues = append(labelValues, extraLabelValues...)
	} else {
		labelValues = extraLabelValues
	}

	metric := prometheus.MustNewConstMetric(desc, valueType, value, labelValues...)

	return metric
}
