package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/me/dkg-node/jsonrpc"
	"github.com/sirupsen/logrus"
)

func (s *JRPCApi) Assign(r *http.Request, args *AssignRequest, result *AssignResponse) error {
	if args.VerifierID == "" {
		return &jsonrpc.Error{Code: jsonrpc.ErrorCodeInvalidParams, Message: jsonrpc.InputError, Data: "VerifierID is empty"}
	}
	// Check verifier is supported
	_, err := s.VerifierService.verifier.Lookup(args.Verifier)
	if err != nil {
		return &jsonrpc.Error{Code: jsonrpc.ErrorCodeInvalidParams, Message: jsonrpc.InputError, Data: "Verifier not supported"}
	}
	// Check quantity keys is assigned
	LastCreatedIndex := s.ABCIService.ABCIApp.state.LastCreatedIndex
	LastUnassignedIndex := s.ABCIService.ABCIApp.state.LastUnassignedIndex
	if LastCreatedIndex < LastUnassignedIndex+2 {
		return &jsonrpc.Error{Code: jsonrpc.ErrorCodeInternal, Message: jsonrpc.InputError, Data: "System is under heavy load for assignments, please try again later"}
	}
	// Broadcast to BFT network
	tendermintService := s.TendermintService
	assMsg := AssignmentBFTTx{VerifierID: args.VerifierID, Verifier: args.Verifier}
	hash, err := tendermintService.bftRPC.Broadcast(assMsg)
	if err != nil {
		return &jsonrpc.Error{Code: jsonrpc.ErrorCodeInternal, Message: jsonrpc.InternalError, Data: "Unable to broadcast: " + err.Error()}
	}
	logrus.Debugf("BFTWS:, hashstring %v", hash)
	// Query
	queryData, _ := json.Marshal(assMsg)
	res, err := tendermintService.bftRPC.client.ABCIQuery(tendermintService.ctx, "GetIndexes", []byte(queryData))
	if err != nil {
		return &jsonrpc.Error{Code: jsonrpc.ErrorCodeInternal, Message: jsonrpc.InternalError, Data: "Failed to check if email exists after assignment: " + err.Error()}
	}
	if res.Response.Code > 0 {
		return &jsonrpc.Error{Code: jsonrpc.ErrorCodeInternal, Message: jsonrpc.InternalError, Data: "Failed to find by email: " + res.Response.Log}
	}
	result.Key = res.Response.Value
	return nil
}

func (s *JRPCApi) Lookup(r *http.Request, args *LookupRequest, result *LookupResponse) error {
	if args.VerifierID == "" {
		return &jsonrpc.Error{Code: jsonrpc.ErrorCodeInvalidParams, Message: jsonrpc.InputError, Data: "VerifierID is empty"}
	}

	return nil
}

func (s *JRPCApi) Commitment(r *http.Request, args *CommitmentRequest, result *CommitmentResponse) error {
	if args.VerifierID == "" {
		return &jsonrpc.Error{Code: jsonrpc.ErrorCodeInvalidParams, Message: jsonrpc.InputError, Data: "VerifierID is empty"}
	}
	return nil
}

func (s *JRPCApi) RetrieveShare(r *http.Request, args *RetrieveRequest, result *RetrieveResponse) error {
	if args.VerifierID == "" {
		return &jsonrpc.Error{Code: 32602, Message: jsonrpc.InputError, Data: "VerifierID is empty"}
	}
	return nil
}

func (s *JRPCApi) Add(r *http.Request, args *AddRequest, result *AddResponse) error {
	fmt.Printf("args: %v\n", args)
	if args.A <= 0 || args.B <= 0 {
		return &jsonrpc.Error{Code: 32602, Message: jsonrpc.InputError, Data: "VerifierID is empty"}
	}
	result.Total = args.A + args.B
	return nil
}
