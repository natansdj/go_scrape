package metric

import (
	"github.com/natansdj/go_scrape/status"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "go_scrape_"

// Metrics implements the prometheus.Metrics interface and
// exposes go_scrape metrics for prometheus
type Metrics struct {
	TotalPushCount *prometheus.Desc
	QueueUsage     *prometheus.Desc
	GetQueueUsage  func() int
}

var getGetQueueUsage = func() int { return 0 }

// NewMetrics returns a new Metrics with all prometheus.Desc initialized
func NewMetrics(c ...func() int) Metrics {
	m := Metrics{
		TotalPushCount: prometheus.NewDesc(
			namespace+"total_push_count",
			"Number of push count",
			nil, nil,
		),
		QueueUsage: prometheus.NewDesc(
			namespace+"queue_usage",
			"Length of internal queue",
			nil, nil,
		),
		GetQueueUsage: getGetQueueUsage,
	}

	if len(c) > 0 {
		m.GetQueueUsage = c[0]
	}

	return m
}

// Describe returns all possible prometheus.Desc
func (c Metrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.TotalPushCount
	ch <- c.QueueUsage
}

// Collect returns the metrics with values
func (c Metrics) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.TotalPushCount,
		prometheus.CounterValue,
		float64(status.StatStorage.GetTotalCount()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.QueueUsage,
		prometheus.GaugeValue,
		float64(c.GetQueueUsage()),
	)
}
