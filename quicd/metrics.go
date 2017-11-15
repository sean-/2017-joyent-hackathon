package main

import (
	"os"
	"os/signal"
	"syscall"

	cgm "github.com/circonus-labs/circonus-gometrics"
	"github.com/pkg/errors"
)

var metrics *cgm.CirconusMetrics

func initMetrics(circonusAPIKey string) (err error) {
	log := logger
	cmc := &cgm.Config{}
	cmc.Debug = false // set to true for debug messages
	cmc.Log = log

	cmc.CheckManager.API.TokenApp = "quicd"

	// Circonus API Token key (https://login.circonus.com/user/tokens)
	if circonusAPIKey == "" {
		cmc.CheckManager.API.TokenKey = os.Getenv("CIRCONUS_API_TOKEN")
	} else {
		cmc.CheckManager.API.TokenKey = circonusAPIKey
	}

	log.Println("Creating new cgm instance")

	metrics, err = cgm.NewCirconusMetrics(cmc)
	if err != nil {
		logger.Println(err)
		return errors.Wrap(err, "unable to initialize metrics")
	}

	logger.Println("Adding ctrl-c trap")
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Println("Received CTRL-C, flushing outstanding metrics before exit")
		metrics.Flush()
		os.Exit(0)
	}()

	logger.Println("Starting to send metrics")

	return nil
}
