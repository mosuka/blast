package errors

import "errors"

var (
	ErrNotFoundLeader    = errors.New("does not found leader")
	ErrNodeAlreadyExists = errors.New("node already exists")
	ErrNodeDoesNotExist  = errors.New("node does not exist")
	ErrNodeNotReady      = errors.New("node not ready")
	ErrNotFound          = errors.New("not found")
	ErrTimeout           = errors.New("timeout")
	ErrNoUpdate          = errors.New("no update")
	ErrNil               = errors.New("data is nil")
	ErrUnsupportedEvent  = errors.New("unsupported event")
)
