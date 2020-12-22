package main

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	exportersMu       sync.RWMutex
	exporterFactories = make(map[string]func() Exporter)
)

type contextValues string

const (
	endpointScrapeDuration contextValues = "endpointScrapeDuration"
	endpointUpMetric       contextValues = "endpointUpMetric"
)

//RegisterExporter makes an exporter available by the provided name.
func RegisterExporter(name string, f func() Exporter) {
	exportersMu.Lock()
	defer exportersMu.Unlock()
	if f == nil {
		panic("exporterFactory is nil")
	}
	exporterFactories[name] = f
}

type exporter struct {
	mutex                        sync.RWMutex
	upMetric                     *prometheus.GaugeVec
	endpointUpMetric             *prometheus.GaugeVec
	endpointScrapeDurationMetric *prometheus.GaugeVec
	exporter                     map[string]Exporter
	self                         string
	lastScrapeOK                 bool
}

//Exporter interface for prometheus metrics. Collect is fetching the data and therefore can return an error
type Exporter interface {
	Collect(ctx context.Context, ch chan<- prometheus.Metric) error
	Describe(ch chan<- *prometheus.Desc)
}

func newExporter() *exporter {
	enabledExporter := make(map[string]Exporter)
	for _, e := range config.EnabledExporters {
		if _, ok := exporterFactories[e]; ok {
			enabledExporter[e] = exporterFactories[e]()
		}
	}

	return &exporter{
		upMetric:                     newGaugeVec("up", "Was the last scrape of rabbitmq successful.", nil),
		endpointUpMetric:             newGaugeVec("module_up", "Was the last scrape of zookeeper successful per module.", []string{"module"}),
		endpointScrapeDurationMetric: newGaugeVec("module_scrape_duration_seconds", "Duration of the last scrape in seconds", []string{"module"}),
		exporter:                     enabledExporter,
		lastScrapeOK:                 true, //return true after start. Value will be updated with each scraping
	}
}

func (e *exporter) LastScrapeOK() bool {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	return e.lastScrapeOK
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, ex := range e.exporter {
		ex.Describe(ch)
	}

	e.upMetric.Describe(ch)
	e.endpointUpMetric.Describe(ch)
	e.endpointScrapeDurationMetric.Describe(ch)
	BuildInfo.Describe(ch)
}

// 实现了prometheus client相关接口的exporter，prometheus会调用这个Collect方法
// 在该方法内部，在调用各个注册进enabledExporter的对象的Collect方法，然后将ctx传给它们
func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	start := time.Now()
	allUp := true

	for name, ex := range e.exporter {
		if err := e.collectWithDuration(ex, name, ch); err != nil {
			log.WithError(err).Warn("retrieving " + name + " failed")
			allUp = false
		}
	}

	if allUp {
		gaugeVecWithLabelValues(e.upMetric).Set(1)
	} else {
		gaugeVecWithLabelValues(e.upMetric).Set(0)
	}

	e.lastScrapeOK = allUp
	e.upMetric.Collect(ch)
	e.endpointUpMetric.Collect(ch)
	e.endpointScrapeDurationMetric.Collect(ch)

	BuildInfo.Collect(ch)

	log.WithField("duration", time.Since(start)).Info("Metrics updated")

}

func (e *exporter) collectWithDuration(ex Exporter, name string, ch chan<- prometheus.Metric) error {
	// 定义传给各个exporter.Collect的上下文(保留以后使用，在本项目中暂时无用)
	ctx := context.Background()

	startModule := time.Now()
	err := ex.Collect(ctx, ch)

	gaugeVecWithLabelValues(e.endpointScrapeDurationMetric, name).Set(time.Since(startModule).Seconds())

	if err != nil {
		gaugeVecWithLabelValues(e.endpointUpMetric, name).Set(0)
	} else {
		gaugeVecWithLabelValues(e.endpointUpMetric, name).Set(1)
	}

	return err
}
