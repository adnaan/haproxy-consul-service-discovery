package main

import (
	"context"
	"log"

	"github.com/takama/k8sapp/pkg/logger"
	stdlog "github.com/takama/k8sapp/pkg/logger/standard"
	"github.com/takama/k8sapp/pkg/system"
)

func main() {
	cfg := &Config{Prefix: "SAMPLE_SERVICE"}
	if err := cfg.Load(cfg.Prefix); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := NewService(ctx, cfg)

	go func() {
		err := service.Listen()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Setup logger
	log := stdlog.New(&logger.Config{
		Level: logger.LevelInfo,
		Time:  true,
		UTC:   true,
	})
	log.Warnf("%s log level is used", logger.LevelDebug.String())
	log.Infof("Service %s listened on %s:%d", cfg.ServiceName+cfg.ServiceID, "localhost", cfg.ServicePort)

	signals := system.NewSignals()
	if err := signals.Wait(log, service); err != nil {
		log.Fatal(err)
	}

	cancel()

}
