package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path"

	"go.uber.org/zap"

	"github.com/MovieStoreGuy/versionist/pkg/goproxy"
	"github.com/MovieStoreGuy/versionist/pkg/manifest"
	"github.com/MovieStoreGuy/versionist/pkg/netrc"
	"github.com/MovieStoreGuy/versionist/pkg/request"
	"github.com/MovieStoreGuy/versionist/pkg/resolve"
)

var (
	configDir = flag.String("config-path", "", "Defines the path to the manifest file")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	reqOps := []request.FactoryFunc{}
	if machines, err := netrc.NewMachinesFromEnvironment(); err != nil {
		log.Warn("Failed to load machine's netrc, unable to provide auth to proxy servers", zap.Error(err))
	} else {
		reqOps = append(reqOps, request.WithNetrcAuthentication(machines))
	}

	m, err := manifest.ReadManifest(ctx, *configDir,
		manifest.WithGoProxyClient(goproxy.NewClient(
			goproxy.WithRequestFactory(
				request.NewRequestFactory(reqOps...),
			),
		)),
	)

	if err != nil {
		log.Panic("Failed to read manifiest", zap.Error(err))
	}

	modifier := resolve.NewModifier(
		path.Dir(*configDir),
		m,
		resolve.WithLogger(log.Named("modifier")),
	)
	if err := modifier.Update(); err != nil {
		log.Error("Failed to modifier go.mod files", zap.Error(err))
	}
	log.Info("Finished processing mod files")
}
