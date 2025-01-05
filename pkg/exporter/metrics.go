package exporter

import (
	"fmt"
	"net/http"

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
		Handler: metricsHandler(names),
	}
	go func() {
		log.WithFields(log.Fields{
			"host": a.config.App.Host,
			"port": a.config.App.Port,
		}).Info("Serving metrics http server")
		if err := a.metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			log.WithError(err).Error()
		}
	}()
}

func metricsHandler(names []string) http.HandlerFunc {
	promHandler := promhttp.Handler()

	ipsetList, _ := netlink.IpsetListAll()
	for _, ipset := range ipsetList {
		// If the ipset name is in the list of names to be exported, or if the list contains "all"
		if lo.Contains(names, ipset.SetName) || lo.Contains(names, "all") {
			IPSetList.WithLabelValues(ipset.SetName).Set(float64(len(ipset.Entries)))
		}
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		promHandler.ServeHTTP(rw, r)
	}
}
