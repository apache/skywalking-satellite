package prometheus

import (
	"errors"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/scrape"
)

// MetadataCache is an adapter to prometheus' scrape.Target  and provide only the functionality which is needed
type MetadataCache interface {
	Metadata(metricName string) (scrape.MetricMetadata, bool)
	SharedLabels() labels.Labels
}

type ScrapeManager interface {
	TargetsAll() map[string][]*scrape.Target
}

type metadataService struct {
	sm ScrapeManager
}

func (s *metadataService) Get(job, instance string) (MetadataCache, error) {
	targetGroup, ok := s.sm.TargetsAll()[job]
	if !ok {
		return nil, errors.New("unable to find a target group with job=" + job)
	}

	// from the same targetGroup, instance is not going to be duplicated
	for _, target := range targetGroup {
		if target.Labels().Get(model.InstanceLabel) == instance {
			return &mCache{target}, nil
		}
	}

	return nil, errors.New("unable to find a target with job=" + job + ", and instance=" + instance)
}

// adapter to get metadata from scrape.Target
type mCache struct {
	t *scrape.Target
}

func (m *mCache) Metadata(metricName string) (scrape.MetricMetadata, bool) {
	return m.t.Metadata(metricName)
}

func (m *mCache) SharedLabels() labels.Labels {
	return m.t.DiscoveredLabels()
}
