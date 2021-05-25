package prometheus

import (
	"context"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
)

// QueueAppender todo appender with queue
type QueueAppender struct {
	Ctx context.Context
	Ms  *metadataService
}

// NewQueueAppender construct QueueAppender
func NewQueueAppender(ctx context.Context, ms *metadataService) *QueueAppender {
	return &QueueAppender{Ctx: ctx, Ms: ms}
}

var _ storage.Appender = (*QueueAppender)(nil)

// always returns 0 to disable label caching
func (qa *QueueAppender) Add(ls labels.Labels, t int64, v float64) (uint64, error) {
	// todo add metrics
	return 0, nil
}

// always returns error since we do not cache
func (qa *QueueAppender) AddFast(_ labels.Labels, _ uint64, _ int64, _ float64) error {
	return storage.ErrNotFound
}

// submit metrics data to consumers
func (qa *QueueAppender) Commit() error {
	// todo send metrics to queue
	return nil
}

func (qa *QueueAppender) Rollback() error {
	return nil
}
