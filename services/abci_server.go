package services

import (
	"context"
	"fmt"
	"os"

	"github.com/me/dkg-node/config"
	abciserver "github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/service"
)

type ABCIService struct {
	ctx     context.Context
	ABCIApp *ABCIApp
	server  service.Service
}

func NewABCIService(services *Services) (*ABCIService, error) {
	abciService := &ABCIService{}

	abciService.ABCIApp, _ = abciService.NewABCIApp()
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	socketAddr := config.GlobalConfig.ABCIServer
	server := abciserver.NewSocketServer(socketAddr, abciService.ABCIApp)
	server.SetLogger(logger)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error starting socket server: %v", err)
		os.Exit(1)
	}
	abciService.server = server
	services.ABCIService = abciService
	return abciService, nil
}

func (a *ABCIService) Name() string {
	return "abci"
}

func (a *ABCIService) OnStart() error {
	a.ABCIApp, _ = a.NewABCIApp()
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	socketAddr := config.GlobalConfig.ABCIServer
	server := abciserver.NewSocketServer(socketAddr, a.ABCIApp)

	server.SetLogger(logger)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error starting socket server: %v", err)
		os.Exit(1)
	}
	// defer server.Stop()
	a.server = server
	return nil
}
func (a *ABCIService) OnStop() error {
	if a.server != nil {
		return a.server.Stop()
	}
	return nil
}
