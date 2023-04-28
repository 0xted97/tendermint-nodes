package services

import (
	"github.com/go-pkgz/jrpc"
)

type Plugin struct {
	*jrpc.Server
}

const (
	AssignMethod = "AssignMethod"
)

type AssignHandle struct {
	jrpc.ServerFn
	state string
}

type AssignRequest struct {
	Verifier   string `json:"verifier"`
	VerifierID string `json:"verifier_id"`
}
type AssignResponse struct {
}

func setUpJRPCHandler() error {
	plugin := Plugin{
		jrpc.NewServer("/jrpc"),
	}
	handle := &AssignHandle{state: "Test"}
	plugin.Add(AssignMethod, handle.Handle)
	plugin.Run(9999)
	return nil
}
