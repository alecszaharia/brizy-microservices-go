package logger

import (
	"os"
	"platform/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
)

func NewLogger(level, sid, name, version string) log.Logger {
	base := log.NewStdLogger(os.Stdout)
	base = log.NewFilter(base, log.FilterLevel(log.ParseLevel(level)))
	return log.With(base,
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", sid,
		"service.name", name,
		"service.version", version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
		"request.id", middleware.RequestID(),
	)
}
