// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//
// Refers to https://github.com/open-telemetry/opentelemetry-collector [Apache-2.0 License]

package prometheus

import (
	"context"
	"errors"
	"math"
	"net"
	"sync/atomic"

	"google.golang.org/protobuf/types/known/timestamppb"

	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"

	"go.opentelemetry.io/collector/obsreport"

	"github.com/prometheus/common/model"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"

	commonpb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	resourcepb "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
)

const (
	portAttr   = "port"
	schemeAttr = "scheme"

	transport  = "http"
	dataformat = "prometheus"
)

var (
	idSeq                 int64
	errTransactionAborted = errors.New("QueueAppender aborted")
	errNoJobInstance      = errors.New("job or instance cannot be found from labels")
	errNoStartTimeMetrics = errors.New("process_start_time_seconds metric is missing")
)

// QueueAppender
type QueueAppender struct {
	id                   int64
	ctx                  context.Context
	isNew                bool
	job                  string
	instance             string
	useStartTimeMetric   bool
	startTimeMetricRegex string
	receiverName         string
	ms                   *metadataService
	node                 *commonpb.Node
	resource             *resourcepb.Resource
	metricBuilder        *metricBuilder
}

// NewQueueAppender construct QueueAppender
func NewQueueAppender(ctx context.Context, useStartTimeMetric bool,
	startTimeMetricRegex string, receiverName string,
	ms *metadataService) *QueueAppender {
	return &QueueAppender{
		id:                   atomic.AddInt64(&idSeq, 1),
		ctx:                  ctx,
		isNew:                true,
		useStartTimeMetric:   useStartTimeMetric,
		startTimeMetricRegex: startTimeMetricRegex,
		receiverName:         receiverName,
		ms:                   ms,
	}
}

var _ storage.Appender = (*QueueAppender)(nil)

// always returns 0 to disable label caching
func (qa *QueueAppender) Add(ls labels.Labels, t int64, v float64) (uint64, error) {
	if math.IsNaN(v) {
		return 0, nil
	}

	select {
	case <-qa.ctx.Done():
		return 0, errTransactionAborted
	default:
	}

	if qa.isNew {
		if err := qa.initAppender(ls); err != nil {
			return 0, err
		}
	}
	return 0, qa.metricBuilder.AddDataPoint(ls, t, v)
}

// always returns error since caching is not supported by Add() function
func (qa *QueueAppender) AddFast(_ labels.Labels, _ uint64, _ int64, _ float64) error {
	return storage.ErrNotFound
}

func (qa *QueueAppender) initAppender(ls labels.Labels) error {
	job, instance := ls.Get(model.JobLabel), ls.Get(model.InstanceLabel)
	if job == "" || instance == "" {
		return errNoJobInstance
	}
	// discover the binding target when this method is called for the first time during a transaction
	mc, err := qa.ms.Get(job, instance)
	if err != nil {
		return err
	}
	qa.job = job
	qa.instance = instance
	qa.node, qa.resource = createNodeAndResource(job, instance, mc.SharedLabels().Get(model.SchemeLabel))
	qa.metricBuilder = newMetricBuilder(mc, qa.useStartTimeMetric, qa.startTimeMetricRegex)
	qa.isNew = false
	return nil
}

// submit metrics data to consumers
func (qa *QueueAppender) Commit() error {
	if qa.isNew {
		return nil
	}

	metrics, _, _, err := qa.metricBuilder.Build()
	if err != nil {
		// Only error by Build() is errNoDataToBuild, with numReceivedPoints set to zero.
		obsreport.EndMetricsReceiveOp(qa.ctx, dataformat, 0, err)
		return err
	}

	if qa.useStartTimeMetric {
		// startTime is mandatory in this case, but may be zero when the
		// process_start_time_seconds metric is missing from the target endpoint.
		if qa.metricBuilder.startTime == 0.0 {
			// Since we are unable to adjust metrics properly, we will drop them
			// and return an error.
			err = errNoStartTimeMetrics
			obsreport.EndMetricsReceiveOp(qa.ctx, dataformat, 0, err)
			return err
		}

		adjustStartTime(qa.metricBuilder.startTime, metrics)
	} else {
		// AdjustMetrics - jobsMap has to be non-nil in this case.
		// Note: metrics could be empty after adjustment, which needs to be checked before passing it on to ConsumeMetrics()
		//todo jobsMap
		//metrics, _ = NewMetricsAdjuster(qa.jobsMap.get(tr.job, tr.instance)).AdjustMetrics(metrics)
	}
	// todo send metrics to queue
	return err
}

func (qa *QueueAppender) Rollback() error {
	return nil
}

func createNodeAndResource(job, instance, scheme string) (*commonpb.Node, *resourcepb.Resource) {
	host, port, err := net.SplitHostPort(instance)
	if err != nil {
		host = instance
	}
	node := &commonpb.Node{
		ServiceInfo: &commonpb.ServiceInfo{Name: job},
		Identifier: &commonpb.ProcessIdentifier{
			HostName: host,
		},
	}
	resource := &resourcepb.Resource{
		Labels: map[string]string{
			portAttr:   port,
			schemeAttr: scheme,
		},
	}
	return node, resource
}

func adjustStartTime(startTime float64, metrics []*metricspb.Metric) {
	startTimeTs := timestampFromFloat64(startTime)
	for _, metric := range metrics {
		switch metric.GetMetricDescriptor().GetType() {
		case metricspb.MetricDescriptor_GAUGE_DOUBLE, metricspb.MetricDescriptor_GAUGE_DISTRIBUTION:
			continue
		default:
			for _, ts := range metric.GetTimeseries() {
				ts.StartTimestamp = startTimeTs
			}
		}
	}
}

func timestampFromFloat64(ts float64) *timestamppb.Timestamp {
	secs := int64(ts)
	nanos := int64((ts - float64(secs)) * 1e9)
	return &timestamppb.Timestamp{
		Seconds: secs,
		Nanos:   int32(nanos),
	}
}
