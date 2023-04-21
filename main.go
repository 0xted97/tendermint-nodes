package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/me/dkg-node/config"
	"github.com/me/dkg-node/services"
)

var path string

func init() {
	flag.StringVar(&path, "config-path", "./config/config.json", "config file")
}

func main() {
	flag.Parse()
	ctx := context.Background()

	// Load config
	globalConfig, err := config.LoadConfig(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v", err)
		os.Exit(1)
	}
	nodeList, err := config.LoadNodeList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v", err)
		os.Exit(1)
	}
	config.GlobalConfig = globalConfig
	config.NodeList = nodeList

	// Initial service
	abciService := services.NewABCIService(ctx)
	p2pService := services.NewP2PService(ctx)
	keyGenService := services.NewKeyGenService(ctx)

	compositeService := services.NewCompositeService(abciService, p2pService, keyGenService)
	// Start all services
	err = compositeService.OnStart()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start composite service:%v", err)
		os.Exit(1)
	}

	// Inject services for service after start
	keyGenService.InjectServices(p2pService, abciService.ABCIApp)

	// Initialize all necessary channels
	nodeListMonitorTicker := time.NewTicker(5 * time.Second)
	establishConnection := make(chan bool)

	go services.NodeListMonitor(nodeListMonitorTicker.C, p2pService, establishConnection)
	<-establishConnection

	keyGenService.GenerateAndSendShares()
	// Stop NodeList monitor ticker
	nodeListMonitorTicker.Stop()

	// Exit the blocking chan
	close(establishConnection)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	err = compositeService.OnStop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not stop composite service:%v", err)
	}
	os.Exit(0)
}
