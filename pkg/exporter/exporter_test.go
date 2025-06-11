package exporter

import (
	"errors"
	"strings"
	"testing"
    "os"

	"github.com/hatamiarash7/ipset-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/vishvananda/netlink"
)

// originalNetlinkIpsetListAll holds the original function
var originalNetlinkIpsetListAll func() ([]netlink.IPSetResult, error)

func TestMain(m *testing.M) {
    // Store original function
    originalNetlinkIpsetListAll = ipsetListAll

    // Run tests
    exitCode := m.Run()

    // Restore original function
    ipsetListAll = originalNetlinkIpsetListAll
    os.Exit(exitCode)
}

func setupTestRegistryAndMetrics() {
    // Create a new registry for each test to avoid interference
    reg := prometheus.NewRegistry()
    prometheus.DefaultRegisterer = reg
    prometheus.DefaultGatherer = reg

    // Re-initialize metrics with the new registry
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

    reg.MustRegister(IPSetEntries)
    reg.MustRegister(IPSetUpdateErrors)
}


func TestUpdateMetrics_Success(t *testing.T) {
    setupTestRegistryAndMetrics()

	cfg := &config.Config{}
	cfg.IPSet.Names = []string{"test_set_1", "test_set_2"}
	// No need for NewExporter for this specific test, directly calling updateMetrics
    // However, updateMetrics is a method on App, so we need an app instance.
    app := &App{config: cfg} // Minimal App instance

	mockIpsets := []netlink.IPSetResult{
		{SetName: "test_set_1", TypeName: "hash:ip", Entries: make([]netlink.IPSetEntry, 5)},
		{SetName: "test_set_2", TypeName: "hash:net", Entries: make([]netlink.IPSetEntry, 10)},
		{SetName: "other_set", TypeName: "bitmap:port", Entries: make([]netlink.IPSetEntry, 3)},
	}
	ipsetListAll = func() ([]netlink.IPSetResult, error) {
		return mockIpsets, nil
	}
	defer func() { ipsetListAll = originalNetlinkIpsetListAll }() // Restore

	app.updateMetrics(cfg.IPSet.Names)

	expectedEntries := `
		# HELP ipset_entries_count The total number of entries in an ipset
		# TYPE ipset_entries_count gauge
		ipset_entries_count{set="test_set_1",type="hash:ip"} 5
		ipset_entries_count{set="test_set_2",type="hash:net"} 10
	`
	if err := testutil.CollectAndCompare(IPSetEntries, strings.NewReader(expectedEntries), "ipset_entries_count"); err != nil {
		t.Errorf("unexpected collecting result for IPSetEntries:\n%s", err)
	}

	if val := testutil.ToFloat64(IPSetUpdateErrors); val != 0 {
		t.Errorf("expected IPSetUpdateErrors to be 0, got %v", val)
	}
}

func TestUpdateMetrics_Success_AllKeyword(t *testing.T) {
    setupTestRegistryAndMetrics()

	cfg := &config.Config{}
	cfg.IPSet.Names = []string{"all"}
    app := &App{config: cfg}

	mockIpsets := []netlink.IPSetResult{
		{SetName: "test_set_1", TypeName: "hash:ip", Entries: make([]netlink.IPSetEntry, 5)},
		{SetName: "other_set", TypeName: "bitmap:port", Entries: make([]netlink.IPSetEntry, 3)},
	}
	ipsetListAll = func() ([]netlink.IPSetResult, error) {
		return mockIpsets, nil
	}
	defer func() { ipsetListAll = originalNetlinkIpsetListAll }()

	app.updateMetrics(cfg.IPSet.Names)

	expectedEntries := `
		# HELP ipset_entries_count The total number of entries in an ipset
		# TYPE ipset_entries_count gauge
		ipset_entries_count{set="other_set",type="bitmap:port"} 3
		ipset_entries_count{set="test_set_1",type="hash:ip"} 5
	`
	if err := testutil.CollectAndCompare(IPSetEntries, strings.NewReader(expectedEntries), "ipset_entries_count"); err != nil {
		t.Errorf("unexpected collecting result for IPSetEntries with 'all' keyword:\n%s", err)
	}
}

func TestUpdateMetrics_NetlinkError(t *testing.T) {
    setupTestRegistryAndMetrics()

	cfg := &config.Config{}
	cfg.IPSet.Names = []string{"test_set_1"}
    app := &App{config: cfg}


	ipsetListAll = func() ([]netlink.IPSetResult, error) {
		return nil, errors.New("netlink list error")
	}
	defer func() { ipsetListAll = originalNetlinkIpsetListAll }()

	app.updateMetrics(cfg.IPSet.Names)

	if val := testutil.ToFloat64(IPSetUpdateErrors); val != 1 {
		t.Errorf("expected IPSetUpdateErrors to be 1, got %v", val)
	}

    // Ensure IPSetEntries is empty or unchanged (has no new metrics)
    // For a fresh collector, expecting no metrics is correct.
    emptyExpected := ``
    if err := testutil.CollectAndCompare(IPSetEntries, strings.NewReader(emptyExpected), "ipset_entries_count"); err != nil {
        // This check might be tricky if metrics persist.
        // A more robust check might be to count the number of metrics collected.
        if c := testutil.CollectAndCount(IPSetEntries); c != 0 {
             t.Errorf("expected IPSetEntries to be empty after error, but found %d metrics. Error from CollectAndCompare: %s", c, err)
        }
    }
}
