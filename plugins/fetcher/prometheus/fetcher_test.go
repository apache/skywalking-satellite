// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
package prometheus

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	"github.com/apache/skywalking-satellite/plugins/fetcher/api"

	promcfg "github.com/prometheus/prometheus/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

func Init() {
	plugin.RegisterPluginCategory(reflect.TypeOf((*api.Fetcher)(nil)).Elem())
	plugin.RegisterPlugin(new(Fetcher))
}

func initFetcher(t *testing.T, cfg plugin.Config) api.Fetcher {
	cfg[plugin.NameField] = Name
	q := api.GetFetcher(cfg)
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
	validateFunc func(t *testing.T, em *v1.SniffData, svc map[string][]float64, bvc map[string][][]*v3.MeterBucketValue)
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
		log.Logger.Fatal("len(jobs) != len(targets), make sure job names are unique")
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
	pCfg, err := promcfg.Load(cfgStr, true, nil)
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

var (
	singleElems = []string{
		"go_threads",
		"http_requests_total",
		"rpc_duration_seconds",
		"rpc_duration_seconds_count",
		"rpc_duration_seconds_sum",
		"http_request_duration_seconds_sum",
		"http_request_duration_seconds_count",
	}

	singleValues = map[string][]float64{
		"go_threads":                          {19, 18},
		"http_requests_total":                 {100, 5, 199, 12},
		"http_request_duration_seconds_sum":   {5000, 5050},
		"http_request_duration_seconds_count": {2500, 2600},
		"rpc_duration_seconds":                {1, 5, 8, 1, 6, 8},
		"rpc_duration_seconds_sum":            {5000, 5002},
		"rpc_duration_seconds_count":          {1000, 1001},
	}

	histogramElems = []string{"http_request_duration_seconds"}

	bucketValues = map[string][][]*v3.MeterBucketValue{
		"http_request_duration_seconds": {
			{
				&v3.MeterBucketValue{Bucket: math.Inf(-1), Count: int64(1000)},
				&v3.MeterBucketValue{Bucket: float64(0.05), Count: int64(1500)},
				&v3.MeterBucketValue{Bucket: float64(0.5), Count: int64(2000)},
				&v3.MeterBucketValue{Bucket: float64(1), Count: int64(2500)},
			},
			{
				&v3.MeterBucketValue{Bucket: math.Inf(-1), Count: int64(1100)},
				&v3.MeterBucketValue{Bucket: float64(0.05), Count: int64(1600)},
				&v3.MeterBucketValue{Bucket: float64(0.5), Count: int64(2100)},
				&v3.MeterBucketValue{Bucket: float64(1), Count: int64(2600)},
			},
		},
	}
)

func TestEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	outputChannel := make(chan *v1.SniffData)
	// 1. setup mock server
	mp, cfg, err := setupMockPrometheus(targets...)
	require.Nilf(t, err, "Failed to create Prometheus config: %v", err)
	defer mp.Close()
	t.Log(cfg)
	fetch(ctx, cfg.ScrapeConfigs, outputChannel)

	singleValueCollection := map[string][]float64{}
	bucketValueCollection := map[string][][]*v3.MeterBucketValue{}

OuterLoop:
	for {
		select {
		case e := <-outputChannel:
			targets[0].validateFunc(t, e, singleValueCollection, bucketValueCollection)
		case <-ctx.Done():
			break OuterLoop
		}
	}
	verifyCollection(t, singleValueCollection, bucketValueCollection)
}

func verifyTarget1(t *testing.T, em *v1.SniffData, svc map[string][]float64, bvc map[string][][]*v3.MeterBucketValue) {
	assert.Equal(t, em.GetMeter().Service, "target1", "Get meter service error")

	if em.GetMeter().GetSingleValue() != nil {
		single := em.GetMeter().GetSingleValue()
		t.Log(single.GetName(), single.GetLabels(), single.GetValue())
		assert.Assert(t, is.Contains(singleElems, single.GetName()), "Mismatch single meter name")
		assert.Assert(t, is.Contains(singleValues[single.GetName()], single.GetValue()), "Mismatch single meter value")
		svc[single.GetName()] = append(svc[single.GetName()], single.GetValue())
	} else {
		histogram := em.GetMeter().GetHistogram()
		t.Log(histogram.GetName(), histogram.GetLabels(), histogram.GetValues())
		assert.Assert(t, is.Contains(histogramElems, histogram.GetName()), "Mismatch histogram meter")
		bvc[histogram.GetName()] = append(bvc[histogram.GetName()], histogram.GetValues())
	}
}

func verifyCollection(t *testing.T, svc map[string][]float64, bvc map[string][][]*v3.MeterBucketValue) {
	for k, v := range singleValues {
		for i, e := range v {
			assert.Equal(t, svc[k][i], e, fmt.Sprintf("%s collection has errors", k))
		}
		t.Logf("%s  collection is OK", k)
	}

	for k, v := range bucketValues {
		for i, e := range v {
			for j, f := range e {
				assert.Equal(t, bvc[k][i][j].Bucket, f.Bucket, fmt.Sprintf("%s collection has errors", k))
				assert.Equal(t, bvc[k][i][j].Count, f.Count, fmt.Sprintf("%s collection has errors", k))
			}
		}
		t.Logf("%s  collection is OK", k)
	}
}

type Config map[string]interface{}

func TestFetcher_ScrapeConfig(t *testing.T) {
	f := &Fetcher{}
	configYaml := f.DefaultConfig()
	t.Log(configYaml)
	// viper
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(strings.NewReader(configYaml))
	assert.NilError(t, err, "cannot read default config in the fetcher plugin")
	cfg := Config{}
	if err := v.MergeConfigMap(cfg); err != nil {
		assert.NilError(t, err, "config merge error")
	}
}
