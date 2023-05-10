package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/me/dkg-node/config"
)

type IEthereumService interface {
	Service
	GetSelfPrivateKey() *ecdsa.PrivateKey
	GetSelfPublicKey() *ecdsa.PublicKey
	SelfSignData(data []byte) ([]byte, error)
}

type EthereumService struct {
	Service
	ctx context.Context

	NodePrivateKey *ecdsa.PrivateKey
	NodePublicKey  *ecdsa.PublicKey
	NodeAddress    common.Address
	NodeIndex      int
	CurrentEpoch   int

	EthCurve elliptic.Curve

	EpochNodeRegister map[int]*NodeRegister // epoch => Node Register => NodeReferences
}

type NodeRegister struct {
	AllConnected bool
	NodeList     []*config.NodeDetail
}

func NewEthereumService(ctx context.Context) *EthereumService {
	return &EthereumService{ctx: ctx}
}

func (es *EthereumService) OnStart() error {
	// Get config
	privateKeyECDSA, err := crypto.HexToECDSA(string(config.GlobalConfig.NodePrivateKey))
	if err != nil {
		return err
	}
	es.NodePrivateKey = privateKeyECDSA
	es.NodePublicKey = &privateKeyECDSA.PublicKey
	es.NodeAddress = crypto.PubkeyToAddress(*es.NodePublicKey)
	es.EthCurve = crypto.S256()

	es.NodeIndex = 0 // TODO: will be set after it has been registered

	return nil
}

func (es *EthereumService) OnStop() error {
	// Perform any cleanup or stop actions if necessary
	return nil
}

func (es *EthereumService) Name() string {
	return "ethereum"
}

func (es *EthereumService) GetSelfPrivateKey() *ecdsa.PrivateKey {
	return es.NodePrivateKey
}

func (es *EthereumService) GetSelfPublicKey() *ecdsa.PublicKey {
	return es.NodePublicKey
}

func (es *EthereumService) SelfSignData(data []byte) ([]byte, error) {
	if es.NodePrivateKey == nil {
		return nil, fmt.Errorf("private key not available")
	}

	signature, err := crypto.Sign(data, es.NodePrivateKey)
	if err != nil {
		return nil, fmt.Errorf("could not sign data: %v", err)
	}

	return signature, nil
}
