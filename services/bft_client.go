package services

import (
	"context"
	"errors"

	"github.com/me/dkg-node/config"
	tmclient "github.com/tendermint/tendermint/rpc/client/http"
)

type BFTClient interface {
	Service
	SendTransaction(tx []byte) error
	Call(method string, params []interface{}) (interface{}, error)
}

type BFTClientService struct {
	ctx    context.Context
	client *tmclient.HTTP
}

func NewBFTClientService(ctx context.Context) *BFTClientService {
	return &BFTClientService{
		ctx: ctx,
	}
}

func (bcs *BFTClientService) OnStart() error {
	client, err := tmclient.New(config.GlobalConfig.SocketServerPort, "/websocket")
	if err != nil {
		return err
	}

	bcs.client = client
	return nil
}

func (bcs *BFTClientService) OnStop() error {
	// Add code to stop the BFT client connection or perform any required cleanup.
	return nil
}

func (bcs *BFTClientService) SendTransaction(tx []byte) error {
	_, err := bcs.client.BroadcastTxSync(bcs.ctx, tx)
	if err != nil {
		return err
	}
	return nil
}

func (bcs *BFTClientService) Call(method string, params []interface{}) (interface{}, error) {
	// Add code to make a call to the BFT client using the specified method and parameters.
	return nil, errors.New("Not implemented")
}
