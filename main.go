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
	"github.com/sirupsen/logrus"
)

var path string

func init() {
	flag.StringVar(&path, "config-path", "./config/config.json", "config file")

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
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
	suite := services.Services{Ctx: ctx}
	suite.ConfigService = globalConfig

	// Initial service
	services.NewEthereumService(&suite)
	services.NewABCIService(&suite)
	services.NewP2PService(&suite)
	services.NewKeyGenService(&suite)
	services.NewVerifierService(&suite)
	services.NewTendermintService(&suite)

	// Inject services for service after start

	// Initialize all necessary channels
	nodeListMonitorTicker := time.NewTicker(5 * time.Second)
	establishConnection := make(chan bool)
	services.TestPublicKey()

	go services.SetUpJRPCHandler()
	go services.NodeListMonitor(nodeListMonitorTicker.C, &suite, establishConnection)
	<-establishConnection
	go services.StartTendermintCore(suite.TendermintService, suite.ConfigService.BasePath+"/tendermint")
	go services.AbciMonitor(suite.TendermintService)
	services.KeyGenStart(suite.KeyGenService)
	// Stop NodeList monitor ticker
	nodeListMonitorTicker.Stop()

	// Exit the blocking chan
	close(establishConnection)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	if err != nil {
		fmt.Fprintf(os.Stderr, "could not stop composite service:%v", err)
	}
	os.Exit(0)
}
