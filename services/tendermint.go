package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/me/dkg-node/config"
	"github.com/sirupsen/logrus"
	tmconfig "github.com/tendermint/tendermint/config"
	tmsecp "github.com/tendermint/tendermint/crypto/secp256k1"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmnode "github.com/tendermint/tendermint/node"
	tmp2p "github.com/tendermint/tendermint/p2p"
	tmprivval "github.com/tendermint/tendermint/privval"

	rpcclient "github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/types"
)

type TendermintService struct {
	ctx             context.Context
	config          *config.Config
	ethereumService *EthereumService

	tmNodeKey *tmp2p.NodeKey

	bftNode *tmnode.Node
	bftRPC  *BFTClientService
}

func NewTendermintService(services *Services) (*TendermintService, error) {
	tendermintService := &TendermintService{
		ctx:             services.Ctx,
		config:          services.ConfigService,
		ethereumService: services.EthereumService,
	}

	services.TendermintService = tendermintService
	err := tendermintService.Initialize()
	if err != nil {
		return nil, err
	}
	return tendermintService, nil
}

func (t *TendermintService) Name() string {
	return "tendermint"
}

func (t *TendermintService) Initialize() error {
	config := t.config
	// mkdir folders for tendermint
	err := os.MkdirAll(config.BasePath+"/tendermint", os.ModePerm)
	if err != nil {
		logrus.WithError(err).Error("could not makedir for tendermint")
	}
	err = os.MkdirAll(config.BasePath+"/tendermint/config", os.ModePerm)
	if err != nil {
		logrus.WithError(err).Error("could not makedir for tendermint/config")
	}
	err = os.MkdirAll(config.BasePath+"/tendermint/data", os.ModePerm)
	if err != nil {
		logrus.WithError(err).Error(("could not makedir for tendermint/data"))
	}
	err = os.Remove(config.BasePath + "/tendermint/data/cs.wal/wal")
	if err != nil {
		logrus.WithError(err).Error("could not remove write ahead log")
	} else {
		logrus.Debug("Removed write ahead log")
	}
	nodeKey, err := os.Open(config.BasePath + "/tendermint/config/node_key.json")
	if err == nil {
		bytVal, _ := ioutil.ReadAll(nodeKey)
		logrus.WithField("NodeKey", string(bytVal)).Debug()
	} else {
		logrus.Debug("Could not find NodeKey")
	}
	privValidatorKey, err := os.Open(config.BasePath + "/tendermint/config/priv_validator_key.json")
	if err == nil {
		bytVal, _ := ioutil.ReadAll(privValidatorKey)
		logrus.WithField("PrivValidatorKey", string(bytVal)).Debug()
	} else {
		logrus.Debug("Could not find PrivValidatorKey")
	}

	// Get default tm base path for generation of nodekey
	defaultConfig := tmconfig.DefaultConfig()
	tmRootPath := config.BasePath + "/tendermint"
	defaultConfig.SetRoot(tmRootPath)
	tmNodeKey, err := tmp2p.LoadOrGenNodeKey(defaultConfig.NodeKeyFile())
	if err != nil {
		logrus.WithError(err).Fatal("NodeKey generation issue")
	}
	t.tmNodeKey = tmNodeKey
	t.bftRPC = nil
	t.bftNode = nil
	// go startTendermintCore(t, tmRootPath)
	// go abciMonitor(t)
	// t.bftRPCWS = nil
	return nil
}

func abciMonitor(t *TendermintService) {
	config := t.config
	interval := time.NewTicker(5 * time.Second)
	for range interval.C {
		bftClient, _ := rpcclient.New(config.BftUri, "/websocket")
		// for subscribe and unsubscribe method calls, use this
		bftClientWS, _ := rpcclient.New(config.BftUri, "/websocket")
		err := bftClientWS.Start()
		if err != nil {
			logrus.WithError(err).Error("could not start the bftWS")
		} else {
			t.bftRPC = NewBFTClientService(t.ctx, bftClient)
			// t.bftRPCWS = bftClientWS
			// t.bftRPCWSStatus = BftRPCWSStatusUp
			break
		}
	}
}

func StartTendermintCore(t *TendermintService, buildPath string) {
	globalConfig := t.config

	defaultTmConfig := tmconfig.DefaultConfig()
	defaultTmConfig.SetRoot(buildPath)

	pv := tmsecp.GenPrivKeySecp256k1(t.ethereumService.NodePrivateKey.D.Bytes())
	pvF := tmprivval.NewFilePV(pv, defaultTmConfig.PrivValidatorKeyFile(), defaultTmConfig.PrivValidatorStateFile())
	pvF.Save()

	// TODO: It will move to smart contract, config.NodeList
	nodeWhitelist, _ := t.ethereumService.NodeListInEpoch(t.ethereumService.CurrentEpoch)
	// Genesis file
	// Set validators from epoch is get from whitelist smart contract
	genDoc := tmtypes.GenesisDoc{
		ChainID:     "main-chain",
		GenesisTime: time.Unix(1578036594, 0),
	}
	var validators []tmtypes.GenesisValidator
	var persistantPeersList []string
	for i, node := range nodeWhitelist {
		//convert pubkey X and Y to tmpubkey
		pubkeyBytes := RawPointToTMPubKey(node.PublicKey)
		val := tmtypes.GenesisValidator{
			Address: pubkeyBytes.Address(),
			PubKey:  pubkeyBytes,
			Power:   1,
			Name:    "" + string(i),
		}
		validators = append(validators, val)
	}
	genDoc.Validators = validators

	defaultTmConfig.P2P.PersistentPeers = strings.Join(persistantPeersList, ",")
	if err := genDoc.SaveAs(defaultTmConfig.GenesisFile()); err != nil {
		logrus.WithError(err).Error("could not save as genesis file")
	}

	// Other config
	defaultTmConfig.ProxyApp = globalConfig.ABCIServer
	defaultTmConfig.Consensus.CreateEmptyBlocks = false // not allow empty block (no transactions)
	defaultTmConfig.BaseConfig.DBBackend = "goleveldb"
	defaultTmConfig.FastSyncMode = false
	defaultTmConfig.RPC.ListenAddress = globalConfig.BftUri
	// Set logger, it should use logrus instead default log of tendermint
	var logger tmlog.Logger
	logger = tmlog.NewTMLogger(logrus.New().Out)

	tmconfig.WriteConfigFile(defaultTmConfig.RootDir+"/config/config.toml", defaultTmConfig)
	//Initial Tendermint Node
	n, err := tmnode.DefaultNewNode(defaultTmConfig, logger)

	if err != nil {
		logrus.WithError(err).Fatal("failed to create tendermint node")
	}
	t.bftNode = n
	logrus.WithField("ListenAddress", defaultTmConfig.P2P.ListenAddress).Info("tendermint P2P Connection")
	logrus.WithField("ListenAddress", defaultTmConfig.RPC.ListenAddress).Info("tendermint Node RPC listening")

	//Start Tendermint Node
	t.bftNode.Start()
}

func RawPointToTMPubKey(pubKey *ecdsa.PublicKey) tmsecp.PubKey {
	//convert pubkey X and Y to tmpubkey
	pubKeyBytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)
	return pubKeyBytes[1:34]
}

func (s *TendermintService) OnStop() error {
	fmt.Println("Stopping Tendermint service...")
	return nil
}
