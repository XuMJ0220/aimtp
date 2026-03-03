package aimtp_watcher

import (
	"context"
	"time"

	"aimtp/internal/aimtp_watcher/biz"
	"aimtp/internal/pkg/log"
)

type Watcher struct {
	biz      biz.WatcherBiz
	interval time.Duration
}

func New(biz biz.WatcherBiz, interval time.Duration) *Watcher {
	return &Watcher{
		biz:      biz,
		interval: interval,
	}
}

func (w *Watcher) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		if err := w.biz.RunOnce(ctx); err != nil {
			log.W(ctx).Errorw("Watcher run failed", "err", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
