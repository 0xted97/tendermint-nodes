package services

import (
	"encoding/json"
	"fmt"

	"github.com/go-pkgz/jrpc"
)

func (h *AssignHandle) Handle(id uint64, params json.RawMessage) jrpc.Response {
	fmt.Printf("h.state: %v\n", h.state)
	var data AssignRequest
	err := json.Unmarshal(params, &data)
	if err != nil {
		return jrpc.EncodeResponse(id, nil, err)
	}
	return jrpc.EncodeResponse(id, "hello, it works", nil)
}
