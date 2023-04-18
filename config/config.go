package config

import (
	"encoding/json"
	"io/ioutil"
)

var GlobalConfig *Config

type Config struct {
	HttpServerPort   string `json:"httpServerPort" env:"HTTP_SERVER_PORT"`
	P2PAddress       string `json:"p2pAddress" env:"P2P_ADDRESS"`
	SocketServerPort string `json:"socketServerPort" env:"SOCKET_SERVER_PORT"`
	DatabasePath     string `json:"databasePath" env:"DATABASE_PATH"`

	NodePrivateKey string `json:"nodePrivateKey" env:"NODE_PRIVATE_KEY"`
}

func LoadConfig(path string) (*Config, error) {
	// Read the file content
	data, err := ioutil.ReadFile("./config/config.json")
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
