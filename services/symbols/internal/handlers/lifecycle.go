package handlers

import (
	"symbols/internal/biz"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"
)

// NewLifecycleEventHandler creates a new lifecycle event handler
func NewLifecycleEventHandler(symbolUC biz.SymbolUseCase, logger log.Logger) *LifecycleEventHandler {
	return &LifecycleEventHandler{
		logger:   log.NewHelper(logger),
		symbolUC: symbolUC,
	}
}

type LifecycleEventHandler struct {
	logger   *log.Helper
	symbolUC biz.SymbolUseCase
}

func (h *LifecycleEventHandler) Handle(msg *message.Message) error {
	ctx := msg.Context()
	// Extract correlation ID for tracing
	correlationID := msg.Metadata.Get("correlation_id")
	h.logger.WithContext(ctx).Infof(
		"Processing lifecycle event - msgID: %s, correlationID: %s",
		msg.UUID,
		correlationID,
	)

	// Business logic here

	return nil
}
