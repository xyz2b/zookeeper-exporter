package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "zookeeper"
)

var (
	extraLabelNames = make([]string, len(config.ExtraLabels) + 1)
	extraLabelValues = make([]string, len(config.ExtraLabels) + 1)
)

func initExtraLabels() {
	extraLabelNames = append(extraLabelNames, "node")
	extraLabelValues = append(extraLabelValues, config.ZkHost)

	if config.ExtraLabels != nil {
		for _, extraLabel := range config.ExtraLabels {
			for k, v := range extraLabel {
				extraLabelNames = append(extraLabelNames, k)
				extraLabelValues = append(extraLabelValues, v)
			}
		}
	}
}

func newGaugeVec(metricName string, docString string, labels []string) *prometheus.GaugeVec {
	if labels != nil {
		labels = append(labels, extraLabelNames...)
	} else {
		labels = extraLabelNames
	}

	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metricName,
			Help:      docString,
		},
		labels,
	)
}

func newDesc(metricName string, docString string, labels []string) *prometheus.Desc {
	if labels != nil {
		labels = append(labels, extraLabelNames...)
	} else {
		labels = extraLabelNames
	}

	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", metricName),
		docString,
		labels,
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
