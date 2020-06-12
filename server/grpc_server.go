package server

import (
	"math"
	"net"
	"time"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/mosuka/blast/metric"
	"github.com/mosuka/blast/protobuf"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type GRPCServer struct {
	grpcAddress string
	service     *GRPCService
	server      *grpc.Server
	listener    net.Listener

	certificateFile string
	keyFile         string
	commonName      string

	logger *zap.Logger
}

func NewGRPCServer(grpcAddress string, raftServer *RaftServer, logger *zap.Logger) (*GRPCServer, error) {
	return NewGRPCServerWithTLS(grpcAddress, raftServer, "", "", "", logger)
}

func NewGRPCServerWithTLS(grpcAddress string, raftServer *RaftServer, certificateFile string, keyFile string, commonName string, logger *zap.Logger) (*GRPCServer, error) {
	grpcLogger := logger.Named("grpc")

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(math.MaxInt64),
		grpc.MaxSendMsgSize(math.MaxInt64),
		grpc.StreamInterceptor(
			grpcmiddleware.ChainStreamServer(
				metric.GrpcMetrics.StreamServerInterceptor(),
				grpczap.StreamServerInterceptor(grpcLogger),
			),
		),
		grpc.UnaryInterceptor(
			grpcmiddleware.ChainUnaryServer(
				metric.GrpcMetrics.UnaryServerInterceptor(),
				grpczap.UnaryServerInterceptor(grpcLogger),
			),
		),
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				//MaxConnectionIdle:     0,
				//MaxConnectionAge:      0,
				//MaxConnectionAgeGrace: 0,
				Time:    5 * time.Second,
				Timeout: 5 * time.Second,
			},
		),
	}

	if certificateFile == "" && keyFile == "" {
		logger.Info("disabling TLS")
	} else {
		logger.Info("enabling TLS")
		creds, err := credentials.NewServerTLSFromFile(certificateFile, keyFile)
		if err != nil {
			logger.Error("failed to create credentials", zap.Error(err))
		}
		opts = append(opts, grpc.Creds(creds))
	}

	server := grpc.NewServer(
		opts...,
	)

	service, err := NewGRPCService(raftServer, certificateFile, commonName, logger)
	if err != nil {
		logger.Error("failed to create key value store service", zap.Error(err))
		return nil, err
	}

	protobuf.RegisterIndexServer(server, service)

	// Initialize all metrics.
	metric.GrpcMetrics.InitializeMetrics(server)
	grpc_prometheus.Register(server)

	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		logger.Error("failed to create listener", zap.String("grpc_address", grpcAddress), zap.Error(err))
		return nil, err
	}

	return &GRPCServer{
		grpcAddress:     grpcAddress,
		service:         service,
		server:          server,
		listener:        listener,
		certificateFile: certificateFile,
		keyFile:         keyFile,
		commonName:      commonName,
		logger:          logger,
	}, nil
}

func (s *GRPCServer) Start() error {
	if err := s.service.Start(); err != nil {
		s.logger.Error("failed to start service", zap.Error(err))
	}

	go func() {
		_ = s.server.Serve(s.listener)
	}()

	s.logger.Info("gRPC server started", zap.String("grpc_address", s.grpcAddress))
	return nil
}

func (s *GRPCServer) Stop() error {
	if err := s.service.Stop(); err != nil {
		s.logger.Error("failed to stop service", zap.Error(err))
	}

	//s.server.GracefulStop()
	s.server.Stop()

	s.logger.Info("gRPC server stopped", zap.String("grpc_address", s.grpcAddress))
	return nil
}
