package prometheus

import "github.com/apache/skywalking-satellite/internal/pkg/log"

type MetricConverter struct {
}

func (mc *MetricConverter) convertMetricsToSampleFamily(metrics []string) {
	for _, metric := range metrics {
		log.Logger.Info(metric)
	}

}
