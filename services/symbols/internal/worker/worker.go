// Package worker implements the Watermill message router and worker lifecycle.
package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"
)

type HookFunc func(ctx context.Context) error

type Worker interface {
	Start() HookFunc
	Stop() HookFunc
}

func NewWorker(router *message.Router, logger log.Logger) Worker {
	return &worker{
		logger:       log.NewHelper(logger),
		name:         "watermill-router",
		router:       router,
		closeTimeout: 15 * time.Second,
		done:         make(chan struct{}),
	}
}

type worker struct {
	logger       *log.Helper
	name         string
	router       *message.Router
	closeTimeout time.Duration

	startOnce sync.Once
	stopOnce  sync.Once

	done chan struct{}
	err  error
}

func (w *worker) Start() HookFunc {
	return func(ctx context.Context) error {
		w.startOnce.Do(func() {
			go func() {
				defer close(w.done)
				w.logger.WithContext(ctx).Infof("Starting router %s", w.name)
				if err := w.router.Run(ctx); err != nil {
					w.err = fmt.Errorf("%s: router run: %w", w.name, err)
					w.logger.WithContext(ctx).Errorf("%v", w.err)
				}
			}()
		})

		return nil
	}
}

func (w *worker) Stop() HookFunc {
	return func(ctx context.Context) error {
		var closeErr error
		w.stopOnce.Do(func() {
			stopCtx, cancel := context.WithTimeout(ctx, w.closeTimeout)
			defer cancel()

			w.logger.WithContext(ctx).Infof("Closing router %s", w.name)

			errCh := make(chan error, 1)
			go func() { errCh <- w.router.Close() }()

			select {
			case <-stopCtx.Done():
				w.logger.WithContext(ctx).Errorf("Shutting down %s (timeout: %v)", w.name, w.closeTimeout)
				closeErr = fmt.Errorf("%s: router close timeout: %w", w.name, stopCtx.Err())
			case err := <-errCh:
				if err != nil {
					w.logger.WithContext(ctx).Errorf("%s: router failed to gracefully close: %v", w.name, err)
					closeErr = fmt.Errorf("%s: router close: %w", w.name, err)
				}
			}
		})

		<-w.done
		if closeErr != nil && w.err == nil {
			w.err = closeErr
		}

		return w.err
	}
}
