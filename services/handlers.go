package services

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	"github.com/me/dkg-node/config"
	"github.com/me/dkg-node/jsonrpc"
	"github.com/sirupsen/logrus"
)

const (
	AssignMethod = "AssignMethod"
)

type (
	CustomError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}

	AssignRequest struct {
		Verifier   string `json:"verifier"`
		VerifierID string `json:"verifier_id"`
	}

	AssignResponse struct {
		Index int    `json:"total"`
		Key   []byte `json:"key"`
	}

	LookupRequest struct {
		Verifier   string `json:"verifier"`
		VerifierID string `json:"verifier_id"`
	}

	LookupResponse struct {
	}

	CommitmentRequest struct {
		Verifier        string `json:"verifier"`
		VerifierID      string `json:"verifier_id"`
		MessagePrefix   string `json:"messageprefix"`
		TokenCommitment string `json:"tokencommitment"`
		TempPubX        string `json:"temppubx"`
		TempPubY        string `json:"temppuby"`
	}

	CommitmentResponse struct {
		Signatures interface{} `json:signatures`
	}

	RetrieveRequest struct {
		Verifier   string `json:"verifier"`
		VerifierID string `json:"verifier_id"`
		IDToken    string `json:"id_token"`
	}

	RetrieveResponse struct {
		Shares []string `json:"shares"`
	}

	AddRequest struct {
		A int `json:"A"`
		B int `json:"B"`
	}

	AddResponse struct {
		Total int `json:"total"`
	}
)

type JRPCApi struct {
	state             string
	ABCIService       *ABCIService
	VerifierService   *VerifierService
	EthereumService   *EthereumService
	TendermintService *TendermintService
}

type CustomCodec struct {
	*jsonrpc.Codec
}

func SetUpJRPCHandler(services *Services) error {
	httpPort := config.GlobalConfig.HttpServerPort
	router := mux.NewRouter()
	server := rpc.NewServer()
	server.RegisterCodec(jsonrpc.NewCodec(), "application/json")
	server.RegisterCodec(jsonrpc.NewCodec(), "application/json;charset=UTF-8")

	jrpcApi := &JRPCApi{
		ABCIService:       services.ABCIService,
		VerifierService:   services.VerifierService,
		EthereumService:   services.EthereumService,
		TendermintService: services.TendermintService,
	}
	server.RegisterService(jrpcApi, "")

	router.Handle("/jrpc", server)
	logrus.WithFields(logrus.Fields{
		"Port": httpPort,
	}).Info("Setting up JSON-RPC handler...")
	http.ListenAndServe(":"+httpPort, router)

	return nil
}
