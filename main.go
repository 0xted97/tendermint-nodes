package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/me/dkg-node/config"
	"github.com/me/dkg-node/services"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", "./config/config.json", "config file")
}

func main() {
	// Load config
	globalConfig, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v", err)
		os.Exit(1)
	}
	ctx := context.Background()
	config.GlobalConfig = globalConfig
	// Initial service
	abciService := services.NewABCIService(ctx)
	p2pService := services.NewP2PService(ctx)
	keyGenService := services.NewKeyGenService(ctx, p2pService)

	compositeService := services.NewCompositeService(abciService, p2pService, keyGenService)
	// Start all services
	err = compositeService.OnStart()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start composite service:%v", err)
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	err = compositeService.OnStop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not stop composite service:%v", err)
	}
	os.Exit(0)
}
