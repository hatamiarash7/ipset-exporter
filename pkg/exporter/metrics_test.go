package exporter

import (
	"errors"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCollectionSnapshots(t *testing.T) {
	counts := map[string]uint32{"first": 1000000, "empty": 0, "ignored": 8}
	var listErr error
	c := newCollector([]string{"first", "empty"}, func() (map[string]uint32, error) { return counts, listErr })
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(c)
	c.update()
	assertMetric(t, registry, "ipset_count", "first", 1000000)
	assertMetric(t, registry, "ipset_count", "empty", 0)
	assertMetric(t, registry, "ipset_up", "", 1)
	assertCountSeries(t, registry, 2)
	lastSuccess := c.state.Load().lastSuccess
	if lastSuccess == 0 {
		t.Fatal("successful collection must set its timestamp")
	}
	counts = map[string]uint32{"empty": 0}
	c.update()
	assertCountSeries(t, registry, 1)
	lastSuccess = c.state.Load().lastSuccess
	listErr = errors.New("permission denied")
	c.update()
	assertCountSeries(t, registry, 0)
	assertMetric(t, registry, "ipset_up", "", 0)
	assertMetric(t, registry, "ipset_collection_errors_total", "", 1)
	assertMetric(t, registry, "ipset_last_success_timestamp_seconds", "", lastSuccess)
	listErr = nil
	counts = map[string]uint32{}
	c.update()
	assertMetric(t, registry, "ipset_up", "", 1)
	assertMetric(t, registry, "ipset_collection_errors_total", "", 1)
	assertCountSeries(t, registry, 0)
}

func TestConcurrentScrapesAndIsolatedRegistries(t *testing.T) {
	c := newCollector([]string{"all"}, func() (map[string]uint32, error) {
		return map[string]uint32{"one": 1, "two": 2}, nil
	})
	r := prometheus.NewPedanticRegistry()
	r.MustRegister(c)
	c.update()
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 50 {
				assertCountSeries(t, r, 2)
			}
		})
	}
	for range 50 {
		c.update()
	}
	wg.Wait()
	other := prometheus.NewPedanticRegistry()
	other.MustRegister(newCollector([]string{"all"}, nil))
	assertCountSeries(t, other, 0)
}

func assertMetric(t *testing.T, r *prometheus.Registry, name, set string, want float64) {
	t.Helper()
	families, err := r.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.Metric {
			if set != "" && (len(metric.Label) != 1 || metric.Label[0].GetValue() != set) {
				continue
			}
			got := metric.GetGauge().GetValue()
			if metric.Counter != nil {
				got = metric.GetCounter().GetValue()
			}
			if got != want {
				t.Errorf("%s{%s}: got %v, want %v", name, set, got, want)
			}
			return
		}
	}
	t.Errorf("metric %s{%s} missing", name, set)
}

func assertCountSeries(t *testing.T, r *prometheus.Registry, want int) {
	t.Helper()
	families, err := r.Gather()
	if err != nil {
		t.Error(err)
		return
	}
	got := 0
	for _, family := range families {
		if family.GetName() == "ipset_count" {
			got = len(family.Metric)
		}
	}
	if got != want {
		t.Errorf("got %d count series, want %d", got, want)
	}
}
