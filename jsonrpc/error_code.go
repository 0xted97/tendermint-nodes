package jsonrpc

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
	InternalError = "Internal error"
	InputError    = "Input error"
)
