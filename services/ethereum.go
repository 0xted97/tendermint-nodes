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

type EthereumService interface {
	Service
	GetSelfPrivateKey() *ecdsa.PrivateKey
	GetSelfPublicKey() *ecdsa.PublicKey
	SelfSignData(data []byte) ([]byte, error)
}

type EthereumServiceImpl struct {
	ctx context.Context

	NodePrivateKey *ecdsa.PrivateKey
	NodePublicKey  *ecdsa.PublicKey
	NodeAddress    common.Address
	NodeIndex      int
	CurrentEpoch   int

	EthCurve elliptic.Curve
}

func NewEthereumService(ctx context.Context) *EthereumServiceImpl {
	return &EthereumServiceImpl{}
}

func (es *EthereumServiceImpl) OnStart() error {
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

func (es *EthereumServiceImpl) OnStop() error {
	// Perform any cleanup or stop actions if necessary
	return nil
}

func (es *EthereumServiceImpl) Name() string {
	return "ethereum"
}

func (es *EthereumServiceImpl) GetSelfPrivateKey() *ecdsa.PrivateKey {
	return es.NodePrivateKey
}

func (es *EthereumServiceImpl) GetSelfPublicKey() *ecdsa.PublicKey {
	return es.NodePublicKey
}

func (es *EthereumServiceImpl) SelfSignData(data []byte) ([]byte, error) {
	if es.NodePrivateKey == nil {
		return nil, fmt.Errorf("private key not available")
	}

	signature, err := crypto.Sign(data, es.NodePrivateKey)
	if err != nil {
		return nil, fmt.Errorf("could not sign data: %v", err)
	}

	return signature, nil
}
