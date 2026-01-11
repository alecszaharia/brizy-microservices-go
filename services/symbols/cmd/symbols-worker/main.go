package main

import (
	"flag"
	"os"
	p "platform/logger"
	"symbols/internal/conf/gen"
	"symbols/internal/worker"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
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

func newApp(worker worker.Worker, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.BeforeStart(worker.Start()),
		kratos.AfterStop(worker.Stop()),
	)
}

func main() {
	flag.Parse()

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

	logger := p.NewLogger(bc.Log.Level, id, Name, Version)

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Log, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for a stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
