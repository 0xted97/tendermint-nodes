package services

import (
	"fmt"
	"net/http"

	"github.com/me/dkg-node/jsonrpc"
)

func (s *JRPCApi) Assign(r *http.Request, args *AssignRequest, result *AssignResponse) error {
	if args.VerifierID == "" {
		return &jsonrpc.Error{Code: jsonrpc.ErrorCodeInvalidParams, Message: "Input error", Data: "VerifierID is empty"}
	}
	return nil
}

func (s *JRPCApi) Lookup(r *http.Request, args *LookupRequest, result *LookupResponse) error {
	if args.VerifierID == "" {
		return &jsonrpc.Error{Code: jsonrpc.ErrorCodeInvalidParams, Message: "Input error", Data: "VerifierID is empty"}
	}

	return nil
}

func (s *JRPCApi) Commitment(r *http.Request, args *CommitmentRequest, result *CommitmentResponse) error {
	if args.VerifierID == "" {
		return &jsonrpc.Error{Code: jsonrpc.ErrorCodeInvalidParams, Message: "Input error", Data: "VerifierID is empty"}
	}
	return nil
}

func (s *JRPCApi) RetrieveShare(r *http.Request, args *RetrieveRequest, result *RetrieveResponse) error {
	if args.VerifierID == "" {
		return &jsonrpc.Error{Code: 32602, Message: "Input error", Data: "VerifierID is empty"}
	}
	return nil
}

func (s *JRPCApi) Add(r *http.Request, args *AddRequest, result *AddResponse) error {
	fmt.Printf("args: %v\n", args)
	if args.A <= 0 || args.B <= 0 {
		return &jsonrpc.Error{Code: 32602, Message: "Input error", Data: "VerifierID is empty"}
	}
	result.Total = args.A + args.B
	return nil
}
