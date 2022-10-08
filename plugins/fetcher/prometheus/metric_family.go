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

package prometheus

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/prometheus/common/model"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/textparse"
	"github.com/prometheus/prometheus/scrape"
)

type MetricFamily interface {
	Add(metricName string, ls labels.Labels, t int64, v float64) error
	IsSameFamily(metricName string) bool
	// to OTLP metrics
	// will return 1. metricspb.Metric with timeseries 2. counter all of timeseries 3. count dropped timeseries
	ToMetric() []*v3.MeterData
}

type metricFamily struct {
	name             string
	mtype            textparse.MetricType
	mc               MetadataCache
	labelKeys        map[string]bool
	labelKeysOrdered []string
	metadata         *scrape.MetricMetadata
	groupOrders      map[string]int
	groups           map[string]*metricGroup
}

type metricGroup struct {
	family       *metricFamily
	name         string
	ts           int64
	ls           labels.Labels
	hasCount     bool
	count        float64
	hasSum       bool
	sum          float64
	value        float64
	complexValue []*dataPoint
}

type dataPoint struct {
	value    float64
	boundary float64
}

func normalizeMetricName(name string) string {
	for _, s := range trimmableSuffixes {
		if strings.HasSuffix(name, s) && name != s {
			return strings.TrimSuffix(name, s)
		}
	}
	return name
}

func newMetricFamily(metricName string, mc MetadataCache) MetricFamily {
	familyName := normalizeMetricName(metricName)
	// lookup metadata based on familyName
	metadata, ok := mc.Metadata(familyName)
	if !ok && metricName != familyName {
		// use the original metricName as metricFamily
		familyName = metricName
		// perform a 2nd lookup with the original metric name. it can happen if there's a metric which is not histogram
		// or summary, but ends with one of those _count/_sum suffixes
		metadata, ok = mc.Metadata(metricName)
		// still not found, this can happen when metric has no TYPE HINT
		if !ok {
			metadata.Metric = familyName
			metadata.Type = textparse.MetricTypeUnknown
		}
	}

	return &metricFamily{
		name:             familyName,
		mtype:            metadata.Type,
		mc:               mc,
		labelKeys:        make(map[string]bool),
		labelKeysOrdered: make([]string, 0),
		metadata:         &metadata,
		groupOrders:      make(map[string]int),
		groups:           make(map[string]*metricGroup),
	}
}

func (mf *metricFamily) Add(metricName string, ls labels.Labels, t int64, v float64) error {
	groupKey := mf.getGroupKey(ls)
	mg := mf.loadMetricGroupOrCreate(groupKey, ls, t)
	switch mf.mtype {
	case textparse.MetricTypeCounter:
		mg.value = v
	case textparse.MetricTypeGauge:
		mg.value = v
	case textparse.MetricTypeHistogram:
		if strings.HasSuffix(metricName, metricsSuffixCount) {
			mg.hasCount = true
			mg.count = v
			mg.name = strings.ReplaceAll(metricName, metricsSuffixCount, "")
		} else if strings.HasSuffix(metricName, metricsSuffixSum) {
			mg.hasSum = true
			mg.sum = v
			mg.name = strings.ReplaceAll(metricName, metricsSuffixSum, "")
		} else if strings.HasSuffix(metricName, metricsSuffixBucket) {
			boundary, err := getBoundary(mf.mtype, ls)
			if err != nil {
				return err
			}
			mg.complexValue = append(mg.complexValue, &dataPoint{value: v, boundary: boundary})
		}
		mg.ts = t
	case textparse.MetricTypeSummary:
		if strings.HasSuffix(metricName, metricsSuffixCount) {
			mg.hasCount = true
			mg.count = v
			mg.name = strings.ReplaceAll(metricName, metricsSuffixCount, "")
		} else if strings.HasSuffix(metricName, metricsSuffixSum) {
			mg.hasSum = true
			mg.sum = v
			mg.name = strings.ReplaceAll(metricName, metricsSuffixSum, "")
		} else {
			mg.value = v
			mg.name = metricName
		}
		mg.ts = t
	default:
		mg.value = v
		mg.name = metricName
	}
	return nil
}

func getBoundary(metricType textparse.MetricType, lbs labels.Labels) (float64, error) {
	labelName := ""
	switch metricType {
	case textparse.MetricTypeHistogram:
		labelName = model.BucketLabel
	case textparse.MetricTypeSummary:
		labelName = model.QuantileLabel
	default:
		return 0, fmt.Errorf("errNoBoundaryLabel")
	}

	v := lbs.Get(labelName)
	if v == "" {
		return 0, fmt.Errorf("errEmptyBoundaryLabel")
	}

	return strconv.ParseFloat(v, 64)
}

func (mf *metricFamily) convertSummaryToSingleValue(mg *metricGroup) []*v3.MeterData {
	result := make([]*v3.MeterData, 0)
	if mg.hasCount || mg.hasSum {
		if mg.hasCount {
			msv := &v3.MeterSingleValue{
				Name:   mg.name + metricsSuffixCount,
				Labels: mf.convertLabels(mg),
				Value:  mg.count,
			}
			result = append(result, &v3.MeterData{
				Metric:    &v3.MeterData_SingleValue{SingleValue: msv},
				Timestamp: mg.ts,
			})
		}
		if mg.hasSum {
			msv := &v3.MeterSingleValue{
				Name:   mg.name + metricsSuffixSum,
				Labels: mf.convertLabels(mg),
				Value:  mg.sum,
			}
			result = append(result, &v3.MeterData{
				Metric:    &v3.MeterData_SingleValue{SingleValue: msv},
				Timestamp: mg.ts,
			})
		}
	} else {
		msv := &v3.MeterSingleValue{
			Name:   mg.name,
			Labels: mf.convertLabels(mg),
			Value:  mg.value,
		}
		result = append(result, &v3.MeterData{
			Metric:    &v3.MeterData_SingleValue{SingleValue: msv},
			Timestamp: mg.ts,
		})
	}
	return result
}

func (mf *metricFamily) ToMetric() []*v3.MeterData {
	result := make([]*v3.MeterData, 0)
	switch mf.mtype {
	case textparse.MetricTypeSummary:
		for _, mg := range mf.getGroups() {
			result = append(result, mf.convertSummaryToSingleValue(mg)...)
		}
	case textparse.MetricTypeHistogram:
		for _, mg := range mf.getGroups() {
			if mg.hasCount {
				msv := &v3.MeterSingleValue{
					Name:   mg.name + metricsSuffixCount,
					Labels: mf.convertLabels(mg),
					Value:  mg.count,
				}
				result = append(result, &v3.MeterData{
					Metric:    &v3.MeterData_SingleValue{SingleValue: msv},
					Timestamp: mg.ts,
				})
			}
			if mg.hasSum {
				msv := &v3.MeterSingleValue{
					Name:   mg.name + metricsSuffixSum,
					Labels: mf.convertLabels(mg),
					Value:  mg.sum,
				}
				result = append(result, &v3.MeterData{
					Metric:    &v3.MeterData_SingleValue{SingleValue: msv},
					Timestamp: mg.ts,
				})
			}

			bucketMap := make(map[float64]float64)
			for _, dp := range mg.complexValue {
				bucketMap[dp.boundary] = dp.value
			}
			sort.Slice(mg.complexValue, func(i, j int) bool {
				return mg.complexValue[i].boundary < mg.complexValue[j].boundary
			})

			mbs := make([]*v3.MeterBucketValue, 0)
			for index, m := range mg.complexValue {
				if index == 0 {
					mbv := &v3.MeterBucketValue{
						Bucket:             math.Inf(-1),
						Count:              int64(m.value),
						IsNegativeInfinity: true,
					}
					mbs = append(mbs, mbv)
				} else {
					mbv := &v3.MeterBucketValue{
						Bucket: mg.complexValue[index-1].boundary,
						Count:  int64(m.value),
					}
					mbs = append(mbs, mbv)
				}
			}
			mh := &v3.MeterHistogram{
				Name:   mf.name,
				Labels: mf.convertLabels(mg),
				Values: mbs,
			}
			result = append(result, &v3.MeterData{
				Metric: &v3.MeterData_Histogram{
					Histogram: mh,
				},
				Timestamp: mg.ts,
			})
		}
	default:
		for _, mg := range mf.getGroups() {
			msv := &v3.MeterSingleValue{
				Name:   mf.name,
				Labels: mf.convertLabels(mg),
				Value:  mg.value,
			}
			result = append(result, &v3.MeterData{
				Metric: &v3.MeterData_SingleValue{SingleValue: msv},
				// job, instance will be added in QueueAppender
				Timestamp: mg.ts,
			})
		}
	}
	return result
}

func (mf *metricFamily) convertLabels(mg *metricGroup) []*v3.Label {
	result := make([]*v3.Label, 0)
	for k, v := range mg.ls.Map() {
		if !isUsefulLabel(mf.mtype, k) {
			continue
		}
		label := &v3.Label{
			Name:  k,
			Value: v,
		}
		result = append(result, label)
	}
	return result
}

func (mf *metricFamily) getGroups() []*metricGroup {
	groups := make([]*metricGroup, len(mf.groupOrders))
	for k, v := range mf.groupOrders {
		groups[v] = mf.groups[k]
	}

	return groups
}

func (mf *metricFamily) IsSameFamily(metricName string) bool {
	// trim known suffix if necessary
	familyName := normalizeMetricName(metricName)
	return mf.name == familyName || familyName != metricName && mf.name == metricName
}

func (mf *metricFamily) getGroupKey(ls labels.Labels) string {
	mf.updateLabelKeys(ls)
	return dpgSignature(mf.labelKeysOrdered, ls)
}

func dpgSignature(orderedKnownLabelKeys []string, ls labels.Labels) string {
	sign := make([]string, 0, len(orderedKnownLabelKeys))
	for _, k := range orderedKnownLabelKeys {
		v := ls.Get(k)
		if v == "" {
			continue
		}
		sign = append(sign, k+"="+v)
	}
	return fmt.Sprintf("%#v", sign)
}

func (mf *metricFamily) updateLabelKeys(ls labels.Labels) {
	for _, l := range ls {
		if isUsefulLabel(mf.mtype, l.Name) {
			if _, ok := mf.labelKeys[l.Name]; !ok {
				mf.labelKeys[l.Name] = true
				// use insertion sort to maintain order
				i := sort.SearchStrings(mf.labelKeysOrdered, l.Name)
				labelKeys := mf.labelKeysOrdered
				labelKeys = append(labelKeys, "")
				copy(labelKeys[i+1:], labelKeys[i:])
				labelKeys[i] = l.Name
				mf.labelKeysOrdered = labelKeys
			}
		}
	}
}

func isUsefulLabel(mType textparse.MetricType, labelKey string) bool {
	result := false
	switch labelKey {
	case model.MetricNameLabel:
	case model.InstanceLabel:
		return false // instance name already in metadata
	case model.SchemeLabel:
	case model.MetricsPathLabel:
	case model.JobLabel:
		return false // service name already in metadata
	case model.BucketLabel: // histogram le
		return mType != textparse.MetricTypeHistogram
	case model.QuantileLabel: // summary quantile
		return true
	default:
		result = true
	}
	return result
}

func (mf *metricFamily) loadMetricGroupOrCreate(groupKey string, ls labels.Labels, ts int64) *metricGroup {
	mg, ok := mf.groups[groupKey]
	if !ok {
		mg = &metricGroup{
			family:       mf,
			ts:           ts,
			ls:           ls,
			complexValue: make([]*dataPoint, 0),
		}
		mf.groups[groupKey] = mg
		// maintaining data insertion order is helpful to generate stable/reproducible metric output
		mf.groupOrders[groupKey] = len(mf.groupOrders)
	}
	return mg
}
