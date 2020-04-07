package client

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	client protobuf.IndexClient

	logger *log.Logger
}

func NewGRPCClient(grpc_address string) (*GRPCClient, error) {
	return NewGRPCClientWithContext(grpc_address, context.Background())
}

func NewGRPCClientWithContext(grpc_address string, baseCtx context.Context) (*GRPCClient, error) {
	return NewGRPCClientWithContextTLS(grpc_address, baseCtx, "", "")
}

func NewGRPCClientWithContextTLS(grpcAddress string, baseCtx context.Context, certificateFile string, commonName string) (*GRPCClient, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(math.MaxInt64),
			grpc.MaxCallRecvMsgSize(math.MaxInt64),
		),
		grpc.WithKeepaliveParams(
			keepalive.ClientParameters{
				Time:                1 * time.Second,
				Timeout:             5 * time.Second,
				PermitWithoutStream: true,
			},
		),
	}

	ctx, cancel := context.WithCancel(baseCtx)

	if certificateFile == "" {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else {
		creds, err := credentials.NewClientTLSFromFile(certificateFile, commonName)
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}

	conn, err := grpc.DialContext(ctx, grpcAddress, dialOpts...)
	if err != nil {
		cancel()
		return nil, err
	}

	return &GRPCClient{
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
		client: protobuf.NewIndexClient(conn),
	}, nil
}

func (c *GRPCClient) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}

	return c.ctx.Err()
}

func (c *GRPCClient) Target() string {
	return c.conn.Target()
}

func (c *GRPCClient) LivenessCheck(opts ...grpc.CallOption) (*protobuf.LivenessCheckResponse, error) {
	if resp, err := c.client.LivenessCheck(c.ctx, &empty.Empty{}, opts...); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (c *GRPCClient) ReadinessCheck(opts ...grpc.CallOption) (*protobuf.ReadinessCheckResponse, error) {
	if resp, err := c.client.ReadinessCheck(c.ctx, &empty.Empty{}, opts...); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (c *GRPCClient) Join(req *protobuf.JoinRequest, opts ...grpc.CallOption) error {
	if _, err := c.client.Join(c.ctx, req, opts...); err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) Leave(req *protobuf.LeaveRequest, opts ...grpc.CallOption) error {
	if _, err := c.client.Leave(c.ctx, req, opts...); err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) Node(opts ...grpc.CallOption) (*protobuf.NodeResponse, error) {
	if resp, err := c.client.Node(c.ctx, &empty.Empty{}, opts...); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (c *GRPCClient) Cluster(opts ...grpc.CallOption) (*protobuf.ClusterResponse, error) {
	if resp, err := c.client.Cluster(c.ctx, &empty.Empty{}, opts...); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (c *GRPCClient) Snapshot(opts ...grpc.CallOption) error {
	if _, err := c.client.Snapshot(c.ctx, &empty.Empty{}); err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) Get(req *protobuf.GetRequest, opts ...grpc.CallOption) (*protobuf.GetResponse, error) {
	if resp, err := c.client.Get(c.ctx, req, opts...); err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			return nil, errors.ErrNotFound
		default:
			return nil, err
		}
	} else {
		return resp, nil
	}
}

func (c *GRPCClient) Search(req *protobuf.SearchRequest, opts ...grpc.CallOption) (*protobuf.SearchResponse, error) {
	if resp, err := c.client.Search(c.ctx, req, opts...); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (c *GRPCClient) Set(req *protobuf.SetRequest, opts ...grpc.CallOption) error {
	if _, err := c.client.Set(c.ctx, req, opts...); err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) Delete(req *protobuf.DeleteRequest, opts ...grpc.CallOption) error {
	if _, err := c.client.Delete(c.ctx, req, opts...); err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) BulkIndex(req *protobuf.BulkIndexRequest, opts ...grpc.CallOption) (*protobuf.BulkIndexResponse, error) {
	if resp, err := c.client.BulkIndex(c.ctx, req, opts...); err == nil {
		return resp, nil
	} else {
		return nil, err
	}
}

func (c *GRPCClient) BulkDelete(req *protobuf.BulkDeleteRequest, opts ...grpc.CallOption) (*protobuf.BulkDeleteResponse, error) {
	if resp, err := c.client.BulkDelete(c.ctx, req, opts...); err == nil {
		return resp, nil
	} else {
		return nil, err
	}
}

func (c *GRPCClient) Mapping(opts ...grpc.CallOption) (*protobuf.MappingResponse, error) {
	if resp, err := c.client.Mapping(c.ctx, &empty.Empty{}, opts...); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (c *GRPCClient) Watch(req *empty.Empty, opts ...grpc.CallOption) (protobuf.Index_WatchClient, error) {
	return c.client.Watch(c.ctx, req, opts...)
}

func (c *GRPCClient) Metrics(opts ...grpc.CallOption) (*protobuf.MetricsResponse, error) {
	if resp, err := c.client.Metrics(c.ctx, &empty.Empty{}, opts...); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}
