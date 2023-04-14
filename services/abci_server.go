package services

import (
	"context"
	"flag"
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
	server  *service.Service
}

func NewABCIService(ctx context.Context) *ABCIService {
	return &ABCIService{ctx: ctx}
}

func (a *ABCIService) Name() string {
	return "abci"
}
func (a *ABCIService) OnStart() error {
	a.ABCIApp = a.NewABCIApp()
	flag.Parse()
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	socketAddr := config.GlobalConfig.SocketServerPort
	server := abciserver.NewSocketServer(socketAddr, a.ABCIApp)

	server.SetLogger(logger)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error starting socket server: %v", err)
		os.Exit(1)
	}
	// defer server.Stop()
	return nil
}
func (a *ABCIService) OnStop() error {
	return nil
}
