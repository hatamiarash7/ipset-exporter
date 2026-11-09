package exporter

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// snapshot is immutable after publication, so a scrape cannot observe a partial update.
type snapshot struct {
	counts      map[string]uint32
	up          float64
	duration    float64
	lastSuccess float64
	errors      uint64
}

type collector struct {
	state       atomic.Pointer[snapshot]
	list        func() (map[string]uint32, error)
	names       map[string]struct{}
	all         bool
	count       *prometheus.Desc
	up          *prometheus.Desc
	duration    *prometheus.Desc
	lastSuccess *prometheus.Desc
	errors      *prometheus.Desc
}

func newCollector(names []string, list func() (map[string]uint32, error)) *collector {
	c := &collector{
		list: list, names: make(map[string]struct{}, len(names)),
		count:       prometheus.NewDesc("ipset_count", "Number of entries reported by the kernel for an ipset.", []string{"set"}, nil),
		up:          prometheus.NewDesc("ipset_up", "Whether the most recent ipset collection succeeded.", nil, nil),
		duration:    prometheus.NewDesc("ipset_collection_duration_seconds", "Duration of the most recent ipset collection.", nil, nil),
		lastSuccess: prometheus.NewDesc("ipset_last_success_timestamp_seconds", "Unix timestamp of the last successful ipset collection, or zero.", nil, nil),
		errors:      prometheus.NewDesc("ipset_collection_errors_total", "Total number of failed ipset collections.", nil, nil),
	}
	for _, name := range names {
		c.names[name] = struct{}{}
		c.all = c.all || name == "all"
	}
	c.state.Store(&snapshot{})
	return c
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range []*prometheus.Desc{c.count, c.up, c.duration, c.lastSuccess, c.errors} {
		ch <- desc
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	s := c.state.Load()
	for name, count := range s.counts {
		ch <- prometheus.MustNewConstMetric(c.count, prometheus.GaugeValue, float64(count), name)
	}
	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, s.up)
	ch <- prometheus.MustNewConstMetric(c.duration, prometheus.GaugeValue, s.duration)
	ch <- prometheus.MustNewConstMetric(c.lastSuccess, prometheus.GaugeValue, s.lastSuccess)
	ch <- prometheus.MustNewConstMetric(c.errors, prometheus.CounterValue, float64(s.errors))
}

// update has one writer: the polling goroutine. Failed or deleted sets never retain stale counts.
func (c *collector) update() {
	start := time.Now()
	counts, err := c.list()
	previous := c.state.Load()
	next := &snapshot{duration: time.Since(start).Seconds(), lastSuccess: previous.lastSuccess, errors: previous.errors}
	if err != nil {
		next.errors++
		log.WithError(err).Error("Failed to collect ipset headers")
	} else {
		next.up = 1
		next.lastSuccess = float64(time.Now().UnixNano()) / 1e9
		next.counts = make(map[string]uint32, len(counts))
		for name, count := range counts {
			if _, selected := c.names[name]; c.all || selected {
				next.counts[name] = count
			}
		}
	}
	c.state.Store(next)
}
