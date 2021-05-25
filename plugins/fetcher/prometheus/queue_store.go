package prometheus

import (
	"context"
	"errors"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
)

var noop = &noopAppender{}

type QueueStore struct {
	ctx                context.Context
	mc                 *metadataService
	useStartTimeMetric bool
	receiverName       string
}

// NewQueueStore construct QueueStore
func NewQueueStore(ctx context.Context, useStartTimeMetric bool, receiverName string) *QueueStore {
	return &QueueStore{
		ctx:                ctx,
		useStartTimeMetric: useStartTimeMetric,
		receiverName:       receiverName,
	}
}

func (qs *QueueStore) SetScrapeManager(scrapeManager *scrape.Manager) {
	if scrapeManager != nil {
		qs.mc = &metadataService{sm: scrapeManager}
	}
}

func (qs *QueueStore) Appender() (storage.Appender, error) {
	return NewQueueAppender(qs.ctx, qs.mc), nil
}

func (qs *QueueStore) Close() error {
	return nil
}

// noopAppender, always return error on any operations
type noopAppender struct{}

func (*noopAppender) Add(labels.Labels, int64, float64) (uint64, error) {
	return 0, errors.New("already stopped")
}

func (*noopAppender) AddFast(labels.Labels, uint64, int64, float64) error {
	return errors.New("already stopped")
}

func (*noopAppender) Commit() error {
	return errors.New("already stopped")
}

func (*noopAppender) Rollback() error {
	return nil
}
