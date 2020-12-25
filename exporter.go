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
	nodeName               contextValues = "node"
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
	confExporter             	 *exporterConf
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
		upMetric:                     newGaugeVec("exporter_up", "Was the last scrape of zookeeper successful.", "node"),
		endpointUpMetric:             newGaugeVec("exporter_module_up", "Was the last scrape of zookeeper successful per module.", "node", "module"),
		endpointScrapeDurationMetric: newGaugeVec("module_scrape_duration_seconds", "Duration of the last scrape in seconds", "node", "module"),
		confExporter:            	  newExporterConf(),
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
	e.confExporter.Describe(ch)
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

	if err := e.collectWithDuration(e.confExporter, "conf", ch); err != nil {
		log.WithError(err).Warn("retrieving overview failed")
		allUp = false
	}

	for name, ex := range e.exporter {
		if err := e.collectWithDuration(ex, name, ch); err != nil {
			log.WithError(err).Warn("retrieving " + name + " failed")
			allUp = false
		}
	}

	if allUp {
		gaugeVecWithLabelValues(e.upMetric, e.confExporter.NodeInfo().Node).Set(1)
	} else {
		gaugeVecWithLabelValues(e.upMetric, e.confExporter.NodeInfo().Node).Set(0)
	}

	e.lastScrapeOK = allUp
	e.upMetric.Collect(ch)
	e.endpointUpMetric.Collect(ch)
	e.endpointScrapeDurationMetric.Collect(ch)

	BuildInfo.Collect(ch)

	log.WithField("duration", time.Since(start)).Info("Metrics updated")

}

func (e *exporter) collectWithDuration(ex Exporter, name string, ch chan<- prometheus.Metric) error {
	// 定义传给各个exporter.Collect的上下文
	ctx := context.Background()
	ctx = context.WithValue(ctx, nodeName, e.confExporter.NodeInfo().Node)

	startModule := time.Now()
	err := ex.Collect(ctx, ch)

	gaugeVecWithLabelValues(e.endpointScrapeDurationMetric, e.confExporter.NodeInfo().Node, name).Set(time.Since(startModule).Seconds())

	if err != nil {
		gaugeVecWithLabelValues(e.endpointUpMetric, e.confExporter.NodeInfo().Node, name).Set(0)
	} else {
		gaugeVecWithLabelValues(e.endpointUpMetric, e.confExporter.NodeInfo().Node, name).Set(1)
	}

	return err
}
