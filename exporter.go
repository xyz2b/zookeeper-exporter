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
	clusterName            contextValues = "cluster"
	totalQueues            contextValues = "totalQueues"
	hostInfo               contextValues = "hostInfo"
	subSystemName          contextValues = "subSystemName"
	subSystemID            contextValues = "subSystemID"
	//extraLabels            contextValues = "extraLabels"
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
}

// 实现了prometheus client相关接口的exporter，prometheus会调用这个Collect方法
// 在该方法内部，在调用各个注册进enabledExporter的对象的Collect方法，然后将ctx传给它们
func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	// 定义传给各个模块Collect的上下文
	ctx := context.Background()

	// 新增: 实例信息(IP:PORT)，来自配置文件的RabbitURL
	ctx = context.WithValue(ctx, hostInfo, config.ZkHost)
	// 新增: 子系统名称，来自配置文件的SubsystemName
	ctx = context.WithValue(ctx, subSystemName, config.SubSystemName)
	// 新增: 子系统ID，来自配置文件的SubsystemID
	ctx = context.WithValue(ctx, subSystemID, config.SubSystemID)
	// 集群名
	ctx = context.WithValue(ctx, clusterName, config.ClusterName)
	// 上报的额外标签信息（附加到所有指标之上）
	//ctx = context.WithValue(ctx, extraLabels, config.ExtraLabels)

	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	start := time.Now()

	for name, ex := range e.exporter {
		if err := e.collectWithDuration(ctx, ex, name, ch); err != nil {
			log.WithError(err).Warn("retrieving " + name + " failed")
		}
	}

	BuildInfo.Collect(ch)

	log.WithField("duration", time.Since(start)).Info("Metrics updated")

}

func (e *exporter) collectWithDuration(ctx context.Context, ex Exporter, name string, ch chan<- prometheus.Metric) error {
	node := ""
	if n, ok := ctx.Value(nodeName).(string); ok {
		node = n
	}
	cluster := ""
	if n, ok := ctx.Value(clusterName).(string); ok {
		cluster = n
	}

	startModule := time.Now()
	err := ex.Collect(ctx, ch)

	if scrapeDuration, ok := ctx.Value(endpointScrapeDuration).(*prometheus.GaugeVec); ok {
		gaugeVecWithLabelValues(&ctx, scrapeDuration, cluster, node, name).Set(time.Since(startModule).Seconds())
	}

	return err
}
