package services

import (
	"fmt"
	"net/http"
)

func (s *JRPCApi) Assign(r *http.Request, args *AssignRequest, result *AssignResponse) error {
	fmt.Printf("args: %v\n", args)
	if args.VerifierID == "" {
		return &JRPCError{Code: 32602, Message: "Input error", Data: "VerifierID is empty"}
	}
	return nil
}

func (s *JRPCApi) Lookup(r *http.Request, args *LookupRequest, result *LookupResponse) error {
	if args.VerifierID == "" {
		return &JRPCError{Code: 32602, Message: "Input error", Data: "VerifierID is empty"}
	}
	return nil
}

func (s *JRPCApi) Commitment(r *http.Request, args *CommitmentRequest, result *CommitmentResponse) error {
	if args.VerifierID == "" {
		return &JRPCError{Code: 32602, Message: "Input error", Data: "VerifierID is empty"}
	}
	return nil
}

func (s *JRPCApi) RetrieveShare(r *http.Request, args *RetrieveRequest, result *RetrieveResponse) error {
	if args.VerifierID == "" {
		return &JRPCError{Code: 32602, Message: "Input error", Data: "VerifierID is empty"}
	}
	return nil
}

func (s *JRPCApi) Add(r *http.Request, args *AddRequest, result *AddResponse) error {
	fmt.Printf("args: %v\n", args)
	if args.A <= 0 || args.B <= 0 {
		return JRPCError{Code: 3000, Message: "Input error", Data: "A,B invalid"}
	}
	result.Total = args.A + args.B
	return nil
}
