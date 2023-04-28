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
	AssignMethod = "AssignMethod"
)

type (
	JrpcError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	AssignHandle struct {
	}

	AssignRequest struct {
		Verifier   string `json:"verifier"`
		VerifierID string `json:"verifier_id"`
	}

	AssignResponse struct {
		Status bool `json:"status"`
		Total  int  `json:"total"`
	}
)

func (e JrpcError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

type Request struct {
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
	ID     uint64      `json:"id"`
}

// Response encloses result and error received from remote server
type Response struct {
	Result interface{} `json:"result,omitempty"`
	ID     uint64      `json:"id"`
}

func (r Response) JRPCResponseSuccess() {

}

func (r Response) JRPCResponseError() {

}

type MyService struct {
	state string
}

type Args struct {
	A int `json:"A"`
	B int `json:"B"`
}

func (s *MyService) Add(r *http.Request, args *Args, result *AssignResponse) error {
	if args.A <= 0 || args.B <= 0 {
		return JrpcError{Code: 3000, Message: "Invalid"}
	}
	result.Total = args.A + args.B
	return nil
}

func (s *MyService) Ping(r *http.Request, args *struct{}, result *Response) error {
	return nil
}

func SetUpJRPCHandler() error {
	httpPort := config.GlobalConfig.HttpServerPort
	fmt.Println("Setting up JSON-RPC handler...")

	router := mux.NewRouter()
	server := rpc.NewServer()

	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	myService := &MyService{state: "Test ne"}
	server.RegisterService(myService, "")

	router.Handle("/jrpc", server)
	http.ListenAndServe(":"+httpPort, router)

	return nil
}
