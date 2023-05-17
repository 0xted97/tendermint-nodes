package services

import (
	"crypto/ecdsa"
	"reflect"

	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
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

type AssignmentBFTTx struct {
	Verifier   string
	VerifierID string
}

// mapping of name of struct to id
var bftTxs = map[string]byte{
	getType(AssignmentBFTTx{}): byte(1),
	// getType(keygennofsm.KeygenMessage{}): byte(2),
	// getType(pss.PSSMessage{}):            byte(3),
	// getType(mapping.MappingMessage{}):    byte(4),
	// getType(dealer.Message{}):            byte(5),
}

func (wrapper *DefaultBFTTxWrapper) PrepareBFTTx(bftTx interface{}, ethereumService *EthereumService) ([]byte, error) {
	// Implement the logic to prepare the BFT transaction
	// type byte
	msgType, ok := bftTxs[getType(bftTx)]
	if !ok {
		return nil, fmt.Errorf("Msg type does not exist for BFT: %s ", getType(bftTx))
	}
	wrapper.MsgType = msgType
	// Generate nonce
	nonce, err := rand.Int(rand.Reader, secp256k1.S256().N)
	if err != nil {
		return nil, fmt.Errorf("Could not generate random number")
	}
	wrapper.Nonce = uint32(nonce.Int64())
	// Set public key this node, to the rest of node can verify signature

	wrapper.PubKey.X = ethereumService.GetSelfPublicKey().X
	wrapper.PubKey.Y = ethereumService.GetSelfPublicKey().Y
	bftRaw, err := json.Marshal(bftTx)
	if err != nil {
		return nil, err
	}
	wrapper.BFTTx = bftRaw

	// Sign data by private key, sign DefaultBFTTxWrapper without signature
	data := wrapper.GetSerializedBody()
	signature, err := ethereumService.SelfSignData(data)
	if err != nil {
		return nil, err
	}
	wrapper.Signature = signature

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
	err := json.Unmarshal(txBytes, &wrapper.BFTTx)
	if err != nil {
		return errors.New("could not decode BFT transaction")
	}
	return nil
}

func (wrapper DefaultBFTTxWrapper) GetSerializedBody() []byte {
	wrapper.Signature = nil
	bin, err := json.Marshal(wrapper)
	if err != nil {
		logrus.Errorf("could not GetSerializedBody bfttx, %v", err)
	}
	return bin
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

type BFTClient interface {
	Service
	SendTransaction(tx []byte) error
	Call(method string, params []interface{}) (interface{}, error)
}

type BFTClientService struct {
	ctx    context.Context
	client *tmclient.HTTP

	ethereumService *EthereumService
}

func NewBFTClientService(ctx context.Context, client *tmclient.HTTP) *BFTClientService {

	return &BFTClientService{
		ctx:    ctx,
		client: client,
	}
}

func (bcs *BFTClientService) Name() string {
	return "bft_client"
}

func (bcs *BFTClientService) Broadcast(tx interface{}) ([]byte, error) {
	var wrapper DefaultBFTTxWrapper
	preparedTx, err := wrapper.PrepareBFTTx(tx, bcs.ethereumService)
	if err != nil {
		logrus.WithError(err).Error("Failed prepare BFT Tx")
		return nil, err
	}
	// Broadcast
	response, err := bcs.client.BroadcastTxSync(bcs.ctx, preparedTx)
	if err != nil {
		return nil, err
	}
	if response.Code != 0 {
		return nil, fmt.Errorf("Could not broadcast, ErrorCode: %v", response.Code)
	}
	return response.Hash, nil
}
