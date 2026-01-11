package data

import (
	"symbols/internal/data/mq"
	"symbols/internal/data/repo"

	"github.com/google/wire"
)

// ProviderSet is Data providers.
var ProviderSet = wire.NewSet(NewDB, NewData, NewTransaction, NewAmqpPublisher, NewAmqpSubscriber, repo.NewSymbolRepo, mq.NewEventPublisher, mq.NewEventSubscriber)
