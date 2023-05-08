package services

import (
	"crypto/ecdsa"
	"math/big"

	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/me/dkg-node/config"
	"github.com/sirupsen/logrus"
	tmclient "github.com/tendermint/tendermint/rpc/client/http"
)

type BFTTxWrapper interface {
	PrepareBFTTx() ([]byte, error)
	DecodeBFTTx([]byte) error
	GetSerializedBody() []byte
}

type DefaultBFTTxWrapper struct {
	BFTTx     []byte          `json:"bft_tx,omitempty"`
	Nonce     uint32          `json:"nonce,omitempty"`
	PubKey    ecdsa.PublicKey `json:"pub_key,omitempty"`
	MsgType   byte            `json:"msg_type,omitempty"`
	Signature []byte          `json:"signature,omitempty"`
}

func (wrapper *DefaultBFTTxWrapper) PrepareBFTTx(bftTx interface{}) ([]byte, error) {
	// Implement the logic to prepare the BFT transaction

	// Generate nonce
	nonce, err := rand.Int(rand.Reader, secp256k1.S256().N)
	if err != nil {
		return nil, fmt.Errorf("Could not generate random number")
	}
	wrapper.Nonce = uint32(nonce.Int64())
	// Set public key this node, to the rest of node can verify signature
	wrapper.PubKey.X = &big.Int{}
	wrapper.PubKey.Y = &big.Int{}

	bftRaw, err := json.Marshal(bftTx)
	if err != nil {
		return nil, err
	}
	wrapper.BFTTx = bftRaw

	// Sign data by private key
	wrapper.Signature = make([]byte, 0)

	// For example, you can marshal the struct to JSON:
	txBytes, err := json.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("could not prepare BFT transaction")
	}

	return txBytes, nil
}

func (wrapper *DefaultBFTTxWrapper) DecodeBFTTx(txBytes []byte) error {
	// Implement the logic to decode the BFT transaction
	// For example, you can unmarshal the JSON into the struct:
	err := json.Unmarshal(txBytes, wrapper)
	if err != nil {
		return errors.New("could not decode BFT transaction")
	}
	return nil
}

func (wrapper DefaultBFTTxWrapper) GetSerializedBody() []byte {
	// Implement the logic to get the serialized body of the BFT transaction
	// For example, you can return the BFTTx field:
	return wrapper.BFTTx
}

type BFTClient interface {
	Service
	SendTransaction(tx []byte) error
	Call(method string, params []interface{}) (interface{}, error)
}

type BFTClientService struct {
	ctx    context.Context
	client *tmclient.HTTP

	compositeService *CompositeService
	ethereumService  *EthereumService
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

func (bcs *BFTClientService) Name() string {
	return "bft_client"
}

func (bcs *BFTClientService) SendTransaction(tx []byte) error {
	_, err := bcs.client.BroadcastTxSync(bcs.ctx, tx)
	if err != nil {
		logrus.WithError(err).Error("Failed to send transaction")
		return err
	}
	return nil
}

func (bcs *BFTClientService) Call(method string, params []interface{}) (interface{}, error) {
	// Add code to make a call to the BFT client using the specified method and parameters.
	return nil, errors.New("Not implemented")
}
