package data

import (
	"symbols/internal/data/repo"

	"github.com/google/wire"
)

// ProviderSet is Data providers.
var ProviderSet = wire.NewSet(
	NewDB,
	NewData,
	NewTransaction,
	NewAMQPPublisher,
	NewAMQPSubscriber,
	repo.NewSymbolRepo,
	NewEventPublisherWithMetrics,
	NewEventSubscriberWithMetrics,
)
