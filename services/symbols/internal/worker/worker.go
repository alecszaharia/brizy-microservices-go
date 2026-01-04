package worker

import (
	"context"
	"fmt"
	"symbols/internal/data/mq"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewRouter, NewWorker, mq.NewSubscriber)

type Worker struct {
	logger       log.Logger
	name         string
	router       *message.Router
	closeTimeout time.Duration

	startOnce sync.Once
	stopOnce  sync.Once

	done chan struct{}
	err  error
}

func NewWorker(router *message.Router, logger log.Logger) *Worker {
	return &Worker{
		logger:       logger,
		name:         "watermill-router",
		router:       router,
		closeTimeout: 15 * time.Second,
		done:         make(chan struct{}),
	}
}

func NewRouter(logger log.Logger) *message.Router {

	stdLogger := watermill.NewStdLogger(false, false)
	router, err := message.NewRouter(message.RouterConfig{}, stdLogger)

	if err != nil {
		panic(err)
	}

	// SignalsHandler will gracefully shut down Router when SIGTERM is received.
	// You can also close the router by just calling `r.Close()`.
	router.AddPlugin(plugin.SignalsHandler)

	// Router level middleware is executed for every message sent to the router
	router.AddMiddleware(
		// CorrelationID will copy the correlation id from the incoming message's metadata to the produced messages
		middleware.CorrelationID,

		// The handler function is retried if it returns an error.
		// After MaxRetries, the message is Nacked, and it's up to the PubSub to resend it.
		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: time.Millisecond * 100,
			Logger:          stdLogger,
		}.Middleware,

		// Recoverer handles panics from handlers.
		// In this case, it passes them as errors to the Retry middleware.
		middleware.Recoverer,
	)

	return router
}

func (w *Worker) WithName(name string) *Worker {
	if name != "" {
		w.name = name
	}
	return w
}

func (w *Worker) WithCloseTimeout(d time.Duration) *Worker {
	if d > 0 {
		w.closeTimeout = d
	}
	return w
}

func (w *Worker) Run(ctx context.Context) error {
	w.startOnce.Do(func() {
		go func() {
			defer close(w.done)
			if err := w.router.Run(ctx); err != nil {
				w.err = fmt.Errorf("%s: router run: %w", w.name, err)
			}
		}()
	})
	return w.err
}

func (w *Worker) Close(ctx context.Context) error {
	w.stopOnce.Do(func() {
		stopCtx, cancel := context.WithTimeout(ctx, w.closeTimeout)
		defer cancel()

		errCh := make(chan error, 1)
		go func() { errCh <- w.router.Close() }()

		select {
		case <-stopCtx.Done():
			if w.err == nil {
				w.err = fmt.Errorf("%s: router close timeout: %w", w.name, stopCtx.Err())
			}
		case err := <-errCh:
			if err != nil && w.err == nil {
				w.err = fmt.Errorf("%s: router close: %w", w.name, err)
			}
		}
	})

	<-w.done
	return w.err
}
