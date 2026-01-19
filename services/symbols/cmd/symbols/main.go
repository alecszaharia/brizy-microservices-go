package main

import (
	"flag"
	"os"
	"platform/build"
	p "platform/logger"
	"symbols/internal/conf/gen"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "symbol-service"
	// Version is the version of the compiled software.
	Version = "1.0"
	// configFile is the config flag.
	configFile string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&configFile, "conf", "configs/config.yaml", "config path, eg: --conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {

	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

var buildInfo = build.NewBuildInfo(Name, Version)

func main() {
	flag.Parse()

	c := config.New(
		config.WithSource(
			env.NewSource(),
			file.NewSource(configFile),
		),
	)
	defer func(c config.Config) {
		err := c.Close()
		if err != nil {
			log.Errorf("failed to close config: %v", err)
		}
	}(c)

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

	app, cleanup, err := wireApp(buildInfo, bc.Server, bc.Data, bc.Log, bc.Metrics, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for a stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
