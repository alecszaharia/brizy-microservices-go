package main

import (
	"context"
	"flag"
	"os"
	"platform/middleware"
	"symbols/internal/conf/gen"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "symbol-worker"
	// Version is the version of the compiled software.
	Version string = "1.0"
	// configFile is the config flag.
	configFile string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&configFile, "conf", "configs/config.yaml", "config path, eg: --conf config.yaml")
}

type Worker interface {
	Run(ctx context.Context) error
	Close(ctx context.Context) error
}

func newApp(runner Worker, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.BeforeStart(func(ctx context.Context) error {
			return runner.Run(ctx)
		}),
		kratos.AfterStop(func(ctx context.Context) error {
			stopCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()
			return runner.Close(stopCtx)
		}),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
		"request.id", middleware.RequestID(),
	)

	c := config.New(
		config.WithSource(
			env.NewSource(),
			file.NewSource(configFile),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	// Config validation
	if err := bc.Validate(); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for a stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
