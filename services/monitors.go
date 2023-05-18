package services

import (
	"fmt"
	"time"

	"github.com/me/dkg-node/config"
)

func NodeListMonitor(tickerChan <-chan time.Time, services *Services, establishConnection chan bool) {
	p2pService := services.P2PService
	ethereumService := services.EthereumService
	epoch := ethereumService.CurrentEpoch
	nodeList, _ := ethereumService.NodeWhitelist(epoch)
	for range tickerChan {
		if _, ok := ethereumService.EpochNodeRegister[epoch]; !ok {
			ethereumService.EpochNodeRegister[epoch] = &NodeRegister{}
		}

		if len(nodeList) != config.GlobalConfig.NumberOfNodes {
			fmt.Println("ethList length not equal to total number of nodes in config...")
			continue
		}

		connectedNodes := make([]*NodeReference, 0)
		for _, node := range nodeList {
			temp, err := p2pService.ConnectToPeer(node)
			if err != nil {
				continue
			}
			connectedNodes = append(connectedNodes, &temp)
		}

		// TODO: currently, cannot connect 3 nodes together, therefore cmt code below
		if len(connectedNodes) != config.GlobalConfig.NumberOfNodes {
			fmt.Println("Not completed connections P2P")
			continue
		}

		fmt.Println("Connected Nodes length is equal to eth list length")
		ethereumService.EpochNodeRegister[epoch].AllConnected = true
		ethereumService.EpochNodeRegister[epoch].NodeList = connectedNodes
		establishConnection <- true
	}
}

func KeyGenStart(k *KeyGenService) error {
	if k.abciApp.state.LastCreatedIndex > 0 {
		fmt.Println("The node already initial share")
		return nil
	}
	fmt.Println("The node is initialing share")
	return k.GenerateAndSendShares()
}
