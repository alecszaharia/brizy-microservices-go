package worker

import (
	"context"
	"fmt"
	"platform/events"
	platform_logger "platform/logger"
	conf "symbols/internal/conf/gen"
	"symbols/internal/handlers"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
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

func NewRouter(cfg *conf.Data, lifecycleHandler *handlers.LifecycleEventHandler, eventPub events.Publisher, eventSub events.Subscriber, logger *platform_logger.WatermillLogger) *message.Router {

	router, err := message.NewRouter(message.RouterConfig{}, logger)

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
			Logger:          logger,
		}.Middleware,

		// Recoverer handles panics from handlers.
		// In this case, it passes them as errors to the Retry middleware.
		middleware.Recoverer,
	)

	router.AddConsumerHandler("events", cfg.Mq.Exchange.Name, eventSub.Unwrap(), lifecycleHandler.Handle)

	return router
}
