package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/me/dkg-node/config"
	"github.com/sirupsen/logrus"
	tmconfig "github.com/tendermint/tendermint/config"
	tmsecp "github.com/tendermint/tendermint/crypto/secp256k1"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmnode "github.com/tendermint/tendermint/node"
	tmp2p "github.com/tendermint/tendermint/p2p"
	rpcclient "github.com/tendermint/tendermint/rpc/client/http"
)

type TendermintService struct {
	ctx             context.Context
	ethereumService EthereumService

	tmNodeKey *tmp2p.NodeKey

	bftNode *tmnode.Node
	bftRPC  *BFTClientService
}

func NewTendermintService(ctx context.Context) *TendermintService {
	return &TendermintService{
		ctx: ctx,
	}
}

func (t *TendermintService) Name() string {
	return "tendermint"
}

func (t *TendermintService) OnStart() error {
	// mkdir folders for tendermint
	err := os.MkdirAll(config.GlobalConfig.BasePath+"/tendermint", os.ModePerm)
	if err != nil {
		logrus.WithError(err).Error("could not makedir for tendermint")
	}
	err = os.MkdirAll(config.GlobalConfig.BasePath+"/tendermint/config", os.ModePerm)
	if err != nil {
		logrus.WithError(err).Error("could not makedir for tendermint/config")
	}
	err = os.MkdirAll(config.GlobalConfig.BasePath+"/tendermint/data", os.ModePerm)
	if err != nil {
		logrus.WithError(err).Error(("could not makedir for tendermint/data"))
	}
	err = os.Remove(config.GlobalConfig.BasePath + "/tendermint/data/cs.wal/wal")
	if err != nil {
		logrus.WithError(err).Error("could not remove write ahead log")
	} else {
		logrus.Debug("Removed write ahead log")
	}
	nodeKey, err := os.Open(config.GlobalConfig.BasePath + "/tendermint/config/node_key.json")
	if err == nil {
		bytVal, _ := ioutil.ReadAll(nodeKey)
		logrus.WithField("NodeKey", string(bytVal)).Debug()
	} else {
		logrus.Debug("Could not find NodeKey")
	}
	privValidatorKey, err := os.Open(config.GlobalConfig.BasePath + "/tendermint/config/priv_validator_key.json")
	if err == nil {
		bytVal, _ := ioutil.ReadAll(privValidatorKey)
		logrus.WithField("PrivValidatorKey", string(bytVal)).Debug()
	} else {
		logrus.Debug("Could not find PrivValidatorKey")
	}

	// Get default tm base path for generation of nodekey
	dftConfig := tmconfig.DefaultConfig()
	tmRootPath := config.GlobalConfig.BasePath + "/tendermint"
	dftConfig.SetRoot(tmRootPath)
	tmNodeKey, err := tmp2p.LoadOrGenNodeKey(dftConfig.NodeKeyFile())
	if err != nil {
		logrus.WithError(err).Fatal("NodeKey generation issue")
	}
	fmt.Printf("tmNodeKey: %v\n", tmNodeKey)
	t.tmNodeKey = tmNodeKey
	t.bftRPC = nil
	t.bftNode = nil
	go startTendermintCore(t, tmRootPath)
	go abciMonitor(t)
	// t.bftRPCWS = nil
	return nil
}

func abciMonitor(t *TendermintService) {
	interval := time.NewTicker(5 * time.Second)
	for range interval.C {
		bftClient, _ := rpcclient.New(config.GlobalConfig.BftUri, "/websocket")
		// for subscribe and unsubscribe method calls, use this
		bftClientWS, _ := rpcclient.New(config.GlobalConfig.BftUri, "/websocket")
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

func startTendermintCore(t *TendermintService, buildPath string) {
	defaultTmConfig := tmconfig.DefaultConfig()
	defaultTmConfig.SetRoot(buildPath)
	defaultTmConfig.ProxyApp = config.GlobalConfig.ABCIServer
	defaultTmConfig.Consensus.CreateEmptyBlocks = false
	defaultTmConfig.BaseConfig.DBBackend = "badgerdb"
	defaultTmConfig.FastSyncMode = false
	defaultTmConfig.RPC.ListenAddress = config.GlobalConfig.BftUri

	// pvEth := t.ethereumService.GetSelfPrivateKey()
	// pvF := tmprivval.NewFilePV(defaultTmConfig.PrivValidatorKeyFile(), defaultTmConfig.PrivValidatorStateFile())
	// fmt.Printf("pvEth: %v\n", pvEth)
	// pvF.Save()
	// pv := tmprivval.LoadFilePV(
	// 	defaultTmConfig.PrivValidatorKeyFile(),
	// 	defaultTmConfig.PrivValidatorStateFile(),
	// )

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

func convertPrivateKey(ethPrivateKey []byte) ([]byte, error) {
	return tmsecp.GenPrivKeySecp256k1(ethPrivateKey), nil
}

func (s *TendermintService) OnStop() error {
	fmt.Println("Stopping Tendermint service...")
	return nil
}
