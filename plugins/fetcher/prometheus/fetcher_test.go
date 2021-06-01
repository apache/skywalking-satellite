package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	fetcher "github.com/apache/skywalking-satellite/plugins/fetcher/api"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"

	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"

	"go.uber.org/zap"

	promcfg "github.com/prometheus/prometheus/config"
)

var logger = zap.NewNop()

func Init() {
	plugin.RegisterPluginCategory(reflect.TypeOf((*fetcher.Fetcher)(nil)).Elem())
	plugin.RegisterPlugin(new(Fetcher))
}

func initFetcher(t *testing.T, cfg plugin.Config) fetcher.Fetcher {
	cfg[plugin.NameField] = Name
	q := fetcher.GetFetcher(cfg)
	if q == nil {
		t.Fatalf("cannot get prometheus-metrics-fetcher from the registry")
	}
	return q
}

type mockPrometheusResponse struct {
	code int
	data string
}

type mockPrometheus struct {
	endpoints   map[string][]mockPrometheusResponse
	accessIndex map[string]*int32
	wg          *sync.WaitGroup
	srv         *httptest.Server
}

func newMockPrometheus(endpoints map[string][]mockPrometheusResponse) *mockPrometheus {
	accessIndex := make(map[string]*int32)
	wg := &sync.WaitGroup{}
	wg.Add(len(endpoints))
	for k := range endpoints {
		v := int32(0)
		accessIndex[k] = &v
	}
	mp := &mockPrometheus{
		wg:          wg,
		accessIndex: accessIndex,
		endpoints:   endpoints,
	}
	srv := httptest.NewServer(mp)
	mp.srv = srv
	return mp
}

func (mp *mockPrometheus) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	iptr, ok := mp.accessIndex[req.URL.Path]
	if !ok {
		rw.WriteHeader(404)
		return
	}
	index := int(*iptr)
	atomic.AddInt32(iptr, 1)
	pages := mp.endpoints[req.URL.Path]
	if index >= len(pages) {
		if index == len(pages) {
			mp.wg.Done()
		}
		rw.WriteHeader(404)
		return
	}
	rw.WriteHeader(pages[index].code)
	_, _ = rw.Write([]byte(pages[index].data))
}

func (mp *mockPrometheus) Close() {
	mp.srv.Close()
}

var srvPlaceHolder = "__SERVER_ADDRESS__"

type testData struct {
	name         string
	pages        []mockPrometheusResponse
	ScrapeConfig *scrapeConfig
	validateFunc func(t *testing.T, em *protocol.Event)
}

func setupMockPrometheus(tds ...*testData) (*mockPrometheus, *promcfg.Config, error) {
	jobs := make([]map[string]interface{}, 0, len(tds))
	endpoints := make(map[string][]mockPrometheusResponse)
	for _, t := range tds {
		metricPath := fmt.Sprintf("/%s/metrics", t.name)
		endpoints[metricPath] = t.pages
		job := make(map[string]interface{})
		job["job_name"] = t.name
		job["metrics_path"] = metricPath
		job["scrape_interval"] = "1s"
		job["static_configs"] = []map[string]interface{}{{"targets": []string{srvPlaceHolder}}}
		jobs = append(jobs, job)
	}

	if len(jobs) != len(tds) {
		logger.Fatal("len(jobs) != len(targets), make sure job names are unique")
	}
	config := make(map[string]interface{})
	config["scrape_configs"] = jobs

	mp := newMockPrometheus(endpoints)
	cfg, err := yaml.Marshal(&config)
	if err != nil {
		return mp, nil, err
	}
	u, _ := url.Parse(mp.srv.URL)
	// update node value (will use for validation)

	cfgStr := strings.ReplaceAll(string(cfg), srvPlaceHolder, u.Host)
	pCfg, err := promcfg.Load(cfgStr)
	return mp, pCfg, err
}

var target1Page1 = `
# HELP go_threads Number of OS threads created
# TYPE go_threads gauge
go_threads 19

# HELP http_requests_total The total number of HTTP requests.
# TYPE http_requests_total counter
http_requests_total{method="post",code="200"} 100
http_requests_total{method="post",code="400"} 5

# HELP http_request_duration_seconds A histogram of the request duration.
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.05"} 1000
http_request_duration_seconds_bucket{le="0.5"} 1500
http_request_duration_seconds_bucket{le="1"} 2000
http_request_duration_seconds_bucket{le="+Inf"} 2500
http_request_duration_seconds_sum 5000
http_request_duration_seconds_count 2500

# HELP rpc_duration_seconds A summary of the RPC duration in seconds.
# TYPE rpc_duration_seconds summary
rpc_duration_seconds{quantile="0.01"} 1
rpc_duration_seconds{quantile="0.9"} 5
rpc_duration_seconds{quantile="0.99"} 8
rpc_duration_seconds_sum 5000
rpc_duration_seconds_count 1000
`

var target1Page2 = `
# HELP go_threads Number of OS threads created
# TYPE go_threads gauge
go_threads 18

# HELP http_requests_total The total number of HTTP requests.
# TYPE http_requests_total counter
http_requests_total{method="post",code="200"} 199
http_requests_total{method="post",code="400"} 12

# HELP http_request_duration_seconds A histogram of the request duration.
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.05"} 1100
http_request_duration_seconds_bucket{le="0.5"} 1600
http_request_duration_seconds_bucket{le="1"} 2100
http_request_duration_seconds_bucket{le="+Inf"} 2600
http_request_duration_seconds_sum 5050
http_request_duration_seconds_count 2600

# HELP rpc_duration_seconds A summary of the RPC duration in seconds.
# TYPE rpc_duration_seconds summary
rpc_duration_seconds{quantile="0.01"} 1
rpc_duration_seconds{quantile="0.9"} 6
rpc_duration_seconds{quantile="0.99"} 8
rpc_duration_seconds_sum 5002
rpc_duration_seconds_count 1001
`

func TestEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	Init()
	initFetcher(t, make(plugin.Config))
	// prepare mock prometheus server
	targets := []*testData{
		{
			name: "target1",
			pages: []mockPrometheusResponse{
				{code: 200, data: target1Page1},
				{code: 500, data: ""},
				{code: 200, data: target1Page2},
			},
			validateFunc: verifyTarget1,
		},
	}
	testEndToEnd(ctx, t, targets)
}

func testEndToEnd(ctx context.Context, t *testing.T, targets []*testData) {
	outputChannel := make(chan *protocol.Event)
	// 1. setup mock server
	mp, cfg, err := setupMockPrometheus(targets...)
	require.Nilf(t, err, "Failed to create Promtheus config: %v", err)
	defer mp.Close()
	t.Log(cfg)
	fetch(ctx, cfg.ScrapeConfigs, outputChannel)
	select {
	case e := <-outputChannel:
		t.Log(e.GetData())
		targets[0].validateFunc(t, e)
	case <-ctx.Done():
		return
	}
}

func verifyTarget1(t *testing.T, em *protocol.Event) {
	assert.Assert(t, em.GetMeter().Service, "prometheus")
}
