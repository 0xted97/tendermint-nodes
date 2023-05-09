package config

import (
	"encoding/json"
	"io/ioutil"
)

var GlobalConfig *Config
var NodeList *([]NodeDetail)

type Config struct {
	HttpServerPort string `json:"httpServerPort" env:"HTTP_SERVER_PORT"`
	P2PAddress     string `json:"p2pAddress" env:"P2P_ADDRESS"`
	BftUri         string `json:"bftUri" env:"BFT_URI"`
	ABCIServer     string `json:"abciServer" env:"SOCKET_SERVER_PORT"`
	DatabasePath   string `json:"databasePath" env:"DATABASE_PATH"`

	EthAddress     string `json:"ethAddress" env:"ETH_ADDRESS"`
	NodePrivateKey string `json:"nodePrivateKey" env:"NODE_PRIVATE_KEY"`

	NumberOfNodes int `json:"numberOfNodes" env:"NUMBER_OF_NODES"`
	KeysPerEpoch  int `json:"keysPerEpoch" env:"KEYS_PER_EPOCH"`

	GoogleClientID     string `json:"googleClientId" env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `json:"googleClientSecret" env:"GOOGLE_CLIENT_SECRET"`

	BasePath string `json:"basePath" env:"BASE_PATH"`
}

func LoadConfig(path string) (*Config, error) {
	// Read the file content
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

type NodeDetail struct {
	Index      int    `json:"index" env:"INDEX"`
	P2PAddress string `json:"p2pAddress" env:"P2P_ADDRESS"`
	EthAddress string `json:"ethAddress" env:"ETH_ADDRESS"`
	Self       bool
}

func LoadNodeList() (*[]NodeDetail, error) {
	// Read the file content
	data, err := ioutil.ReadFile("./config/node-list.json")
	if err != nil {
		return nil, err
	}
	type NodeList struct {
		Nodes []NodeDetail `json:"nodes"`
	}
	var list NodeList
	err = json.Unmarshal(data, &list)
	if err != nil {
		return nil, err
	}
	return &list.Nodes, nil
}
