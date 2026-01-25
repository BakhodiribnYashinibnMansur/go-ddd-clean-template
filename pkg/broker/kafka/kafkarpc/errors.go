package kafkarpc

import "errors"

const (
	Success           = "success"
	ErrInternalServer = "internal_server_error"
	ErrBadHandler     = "bad_handler"
)

var ErrTimeout = errors.New("timeout")
