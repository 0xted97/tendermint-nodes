package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/me/dkg-node/config"
)

type IEthereumService interface {
	Service
	GetSelfPrivateKey() *ecdsa.PrivateKey
	GetSelfPublicKey() *ecdsa.PublicKey
	SelfSignData(data []byte) ([]byte, error)
}

type EthereumService struct {
	ctx    context.Context
	config *config.Config

	NodePrivateKey *ecdsa.PrivateKey
	NodePublicKey  *ecdsa.PublicKey
	NodeAddress    common.Address
	NodeIndex      int
	CurrentEpoch   int

	EthCurve elliptic.Curve

	EpochNodeRegister map[int]*NodeRegister // epoch => Node Register => NodeReferences
}

type NodeReference struct {
	Address         *common.Address
	Index           *big.Int
	PeerID          peer.ID
	PublicKey       *ecdsa.PublicKey
	TMP2PConnection string
	Power           int64
	// P2PConnection   string
	Self bool
}

type NodeRegister struct {
	AllConnected bool
	NodeList     []*NodeReference
}

func NewEthereumService(services *Services) (*EthereumService, error) {
	ethereumService := &EthereumService{
		ctx: services.Ctx, config: services.ConfigService,
	}

	privateKeyECDSA, err := crypto.HexToECDSA(string(config.GlobalConfig.NodePrivateKey))
	if err != nil {
		return nil, err
	}
	ethereumService.NodePrivateKey = privateKeyECDSA
	ethereumService.NodePublicKey = &privateKeyECDSA.PublicKey
	ethereumService.NodeAddress = crypto.PubkeyToAddress(*ethereumService.NodePublicKey)
	ethereumService.EthCurve = crypto.S256()

	ethereumService.EpochNodeRegister = make(map[int]*NodeRegister)
	ethereumService.NodeIndex = 0    // TODO: will be set after it has been registered
	ethereumService.CurrentEpoch = 0 // TODO: will be set after it has been registered
	services.EthereumService = ethereumService
	return ethereumService, nil
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

/*---------------------------------------- Interact contract function -----------------------------------*/
// TODO
func (es *EthereumService) NodeListInEpoch(epoch int) ([]*NodeReference, error) {
	if epoch < 0 {
		return nil, fmt.Errorf("invalid epoch")
	}
	if _, ok := es.EpochNodeRegister[epoch]; !ok {
		return nil, fmt.Errorf("not found")
	}
	// Return eth list in smart contract
	return es.EpochNodeRegister[epoch].NodeList, nil
}

// TODO: Return eth list address in smart contract
func (es *EthereumService) NodeWhitelist(epoch int) ([]common.Address, error) {
	if epoch < 0 {
		return nil, fmt.Errorf("invalid epoch")
	}
	var list []common.Address
	for _, v := range *config.NodeList {
		list = append(list, common.HexToAddress(v.EthAddress))
	}
	return list, nil
}

func (es *EthereumService) NodeDetail(nodeAddress common.Address) (config.NodeDetail, error) {
	for _, v := range *config.NodeList {
		if common.HexToAddress(v.EthAddress).Hex() == nodeAddress.Hex() {
			return v, nil
		}
	}

	return config.NodeDetail{}, fmt.Errorf("not found")
}
