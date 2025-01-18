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

// IPSetList is a gauge vector that holds the number of entries in each ipset
var IPSetList *prometheus.GaugeVec

func (a *App) registerMetrics() {
	IPSetList = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "count",
		Namespace: "ipset",
		Help:      "The total number of ipset entries",
	}, []string{"set"})

	prometheus.Register(IPSetList)
	prometheus.Unregister(collectors.NewGoCollector())
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
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
	ipsetList, err := netlink.IpsetListAll()
	if err != nil {
		log.WithError(err).Error("Failed to list ipsets")
		return
	}

	for _, ipset := range ipsetList {
		// If the ipset name is in the list of names to be exported, or if the list contains "all"
		if lo.Contains(names, ipset.SetName) || lo.Contains(names, "all") {
			IPSetList.WithLabelValues(ipset.SetName).Set(float64(len(ipset.Entries)))
		}
	}
}

func metricsHandler() http.HandlerFunc {
	promHandler := promhttp.Handler()

	return func(rw http.ResponseWriter, r *http.Request) {
		promHandler.ServeHTTP(rw, r)
	}
}
