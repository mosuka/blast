package client

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

type CloudConfig struct {
	Servers     []string
	DialTimeout time.Duration
	DialOptions []grpc.DialOption
	Context     context.Context
}
