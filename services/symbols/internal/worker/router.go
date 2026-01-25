// Package worker implements the Watermill message router and worker lifecycle.
package worker

import (
	"platform/events"
	platform_logger "platform/logger"
	conf "symbols/internal/conf/gen"
	"symbols/internal/handlers"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
)

func NewRouter(cfg *conf.Data, lifecycleHandler *handlers.LifecycleEventHandler, eventSub events.Subscriber, logger *platform_logger.WatermillLogger) *message.Router {

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
