package exporter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

var ipsetListAll = netlink.IpsetListAll

// IPSetEntries is a gauge vector that holds the number of entries in each ipset
var IPSetEntries *prometheus.GaugeVec

// IPSetUpdateErrors is a counter that holds the total number of errors encountered during ipset updates
var IPSetUpdateErrors prometheus.Counter

func (a *App) registerMetrics() {
	IPSetEntries = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "entries_count",
		Namespace: "ipset",
		Help:      "The total number of entries in an ipset",
	}, []string{"set", "type"})

	IPSetUpdateErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "update_errors_total",
		Namespace: "ipset",
		Help:      "The total number of errors encountered during ipset updates.",
	})

	prometheus.MustRegister(IPSetEntries)
	prometheus.MustRegister(IPSetUpdateErrors)
	_ = prometheus.Unregister(collectors.NewGoCollector())
	_ = prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

func (a *App) serveMetrics(names []string) {
	a.metricsServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", a.config.App.Host, a.config.App.Port),
		Handler: metricsHandler(),
	}
	go func() {
		log.WithFields(log.Fields{
			"host": a.config.App.Host,
			"port": a.config.App.Port,
		}).Info("Serving metrics http server")

		// Update metrics before starting the server
		a.updateMetrics(names)

		if err := a.metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			log.WithError(err).Error()
		}
	}()
}

func (a *App) updateMetricsPeriodically(names []string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.updateMetrics(names)
		case <-a.stopChan:
			log.Info("Stopping metrics updater")
			return
		}
	}
}

func (a *App) updateMetrics(names []string) {
	ipsetList, err := ipsetListAll()
	if err != nil {
		log.WithError(err).Error("Failed to list ipsets")
		IPSetUpdateErrors.Inc()
		return
	}

	for _, ipset := range ipsetList {
		// If the ipset name is in the list of names to be exported, or if the list contains "all"
		if lo.Contains(names, ipset.SetName) || lo.Contains(names, "all") {
			log.WithFields(log.Fields{
				"set":     ipset.SetName,
				"type":    ipset.TypeName,
				"entries": len(ipset.Entries),
			}).Debug("Updating metrics")
			IPSetEntries.WithLabelValues(ipset.SetName, ipset.TypeName).Set(float64(len(ipset.Entries)))
		}
	}
}

func metricsHandler() http.HandlerFunc {
	promHandler := promhttp.Handler()

	return func(rw http.ResponseWriter, r *http.Request) {
		promHandler.ServeHTTP(rw, r)
	}
}
