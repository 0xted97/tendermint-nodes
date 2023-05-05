package services

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	"github.com/me/dkg-node/config"
	"github.com/me/dkg-node/jsonrpc"
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
		Status bool `json:"status"`
		Total  int  `json:"total"`
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
	state string
}

type CustomCodec struct {
	*jsonrpc.Codec
}

func SetUpJRPCHandler() error {
	httpPort := config.GlobalConfig.HttpServerPort
	fmt.Println("Setting up JSON-RPC handler...")

	router := mux.NewRouter()
	server := rpc.NewServer()
	server.RegisterCodec(jsonrpc.NewCodec(), "application/json")
	server.RegisterCodec(jsonrpc.NewCodec(), "application/json;charset=UTF-8")

	jrpcApi := &JRPCApi{state: "Test ne"}
	server.RegisterService(jrpcApi, "")

	router.Handle("/jrpc", server)
	http.ListenAndServe(":"+httpPort, router)

	return nil
}
