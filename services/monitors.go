package services

import (
	"fmt"
	"time"

	"github.com/me/dkg-node/config"
)

func NodeListMonitor(tickerChan <-chan time.Time, p2pService *P2PService, establishConnection chan bool) {
	for range tickerChan {
		nodeList := *config.NodeList
		if len(nodeList) != config.GlobalConfig.NumberOfNodes {
			fmt.Println("ethList length not equal to total number of nodes in config...")
			continue
		}

		connectedNodes := make([]*config.NodeDetail, 0)
		for _, node := range nodeList {
			temp, err := p2pService.ConnectToPeer(node)
			if err != nil {
				continue
			}
			connectedNodes = append(connectedNodes, &temp)
		}
		if len(connectedNodes) != config.GlobalConfig.NumberOfNodes {
			fmt.Println("Not completed connections P2P")
			continue
		}
		fmt.Println("Connected Nodes length is equal to eth list length")
		establishConnection <- true
	}
}
