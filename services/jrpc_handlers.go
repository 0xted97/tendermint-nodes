package services

import (
	"encoding/json"

	"github.com/go-pkgz/jrpc"
)

func (h *AssignHandle) Handle(id uint64, params json.RawMessage) jrpc.Response {

	res := AssignResponse{Status: true}
	return jrpc.EncodeResponse(id, res, nil)
}
