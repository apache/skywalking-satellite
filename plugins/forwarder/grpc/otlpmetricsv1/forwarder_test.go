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

package otlpmetricsv1

import (
	"testing"

	metrics "skywalking.apache.org/repo/goapi/proto/opentelemetry/proto/collector/metrics/v1"
	common "skywalking.apache.org/repo/goapi/proto/opentelemetry/proto/common/v1"
	v1 "skywalking.apache.org/repo/goapi/proto/opentelemetry/proto/metrics/v1"
	resource "skywalking.apache.org/repo/goapi/proto/opentelemetry/proto/resource/v1"
)

func strKV(key, value string) *common.KeyValue {
	return &common.KeyValue{
		Key:   key,
		Value: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: value}},
	}
}

func requestWith(attrs ...*common.KeyValue) *metrics.ExportMetricsServiceRequest {
	return &metrics.ExportMetricsServiceRequest{
		ResourceMetrics: []*v1.ResourceMetrics{
			{Resource: &resource.Resource{Attributes: attrs}},
		},
	}
}

func TestGenerateRoutingKeyFollowsMetadataKeyOrder(t *testing.T) {
	f := &Forwarder{metadataKeys: []string{"service.name", "job"}}

	// Same attribute set, different arrival order: the key must be identical
	// and follow the metadataKeys order, not the arrival order.
	first, err := f.generateRoutingKey(requestWith(strKV("job", "my-job"), strKV("service.name", "my-service")))
	if err != nil {
		t.Fatalf("generateRoutingKey returned error: %v", err)
	}
	second, err := f.generateRoutingKey(requestWith(strKV("service.name", "my-service"), strKV("job", "my-job")))
	if err != nil {
		t.Fatalf("generateRoutingKey returned error: %v", err)
	}
	if first != second {
		t.Fatalf("routing key depends on attribute arrival order: %q vs %q", first, second)
	}
	if first != ",my-service,my-job" {
		t.Fatalf("routing key does not follow metadataKeys order: got %q, want %q", first, ",my-service,my-job")
	}
}

func TestGenerateRoutingKeySkipsResourceWithoutRoutingKey(t *testing.T) {
	f := &Forwarder{metadataKeys: []string{"service.name"}}
	req := &metrics.ExportMetricsServiceRequest{
		ResourceMetrics: []*v1.ResourceMetrics{
			{Resource: &resource.Resource{Attributes: []*common.KeyValue{strKV("unrelated", "x")}}},
			{Resource: &resource.Resource{Attributes: []*common.KeyValue{strKV("service.name", "my-service")}}},
		},
	}
	key, err := f.generateRoutingKey(req)
	if err != nil {
		t.Fatalf("generateRoutingKey returned error: %v", err)
	}
	if key != ",my-service" {
		t.Fatalf("got %q, want %q", key, ",my-service")
	}
}

func TestGenerateRoutingKeyNoRoutingKey(t *testing.T) {
	f := &Forwarder{metadataKeys: []string{"service.name"}}
	_, err := f.generateRoutingKey(requestWith(strKV("unrelated", "x")))
	if err == nil {
		t.Fatal("expected error for attributes without any routing key, got nil")
	}
}
