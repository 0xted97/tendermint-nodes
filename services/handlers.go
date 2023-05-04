package services

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/me/dkg-node/config"
)

const (
	// ErrorCodeParse is parse error code.
	ErrorCodeParse = -32700
	// ErrorCodeInvalidRequest is invalid request error code.
	ErrorCodeInvalidRequest = -32600
	// ErrorCodeMethodNotFound is method not found error code.
	ErrorCodeMethodNotFound = -32601
	// ErrorCodeInvalidParams is invalid params error code.
	ErrorCodeInvalidParams = -32602
	// ErrorCodeInternal is internal error code.
	ErrorCodeInternal = -32603
)

const (
	AssignMethod = "AssignMethod"
)

type (
	JRPCError struct {
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
		Verifier   string `json:"verifier"`
		VerifierID string `json:"verifier_id"`
		IdToken    string `json:"id_token"`
	}

	CommitmentResponse struct {
		Signatures interface{} `json:signatures`
	}

	RetrieveRequest struct {
		Verifier   string `json:"verifier"`
		VerifierID string `json:"verifier_id"`
		IdToken    string `json:"id_token"`
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

func (e JRPCError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

type JRPCApi struct {
	state string
}

func SetUpJRPCHandler() error {
	httpPort := config.GlobalConfig.HttpServerPort
	fmt.Println("Setting up JSON-RPC handler...")

	router := mux.NewRouter()
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	jrpcApi := &JRPCApi{state: "Test ne"}
	server.RegisterService(jrpcApi, "")

	router.Handle("/jrpc", server)
	http.ListenAndServe(":"+httpPort, router)

	return nil
}
