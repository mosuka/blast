package server

import (
	"context"
	"math"
	"net"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mosuka/blast/marshaler"
	"github.com/mosuka/blast/protobuf"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

func responseFilter(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	switch resp.(type) {
	case *protobuf.GetResponse:
		w.Header().Set("Content-Type", "application/json")
	case *protobuf.MetricsResponse:
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	default:
		w.Header().Set("Content-Type", marshaler.DefaultContentType)
	}

	return nil
}

type GRPCGateway struct {
	httpAddress string
	grpcAddress string

	cancel   context.CancelFunc
	listener net.Listener
	mux      *runtime.ServeMux

	certificateFile string
	keyFile         string

	corsAllowedMethods []string
	corsAllowedOrigins []string
	corsAllowedHeaders []string

	logger *zap.Logger
}

func NewGRPCGateway(httpAddress string, grpcAddress string, certificateFile string, keyFile string, commonName string, corsAllowedMethods []string, corsAllowedOrigins []string, corsAllowedHeaders []string, logger *zap.Logger) (*GRPCGateway, error) {
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

	baseCtx := context.TODO()
	ctx, cancel := context.WithCancel(baseCtx)

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, new(marshaler.BlastMarshaler)),
		runtime.WithForwardResponseOption(responseFilter),
	)

	if certificateFile == "" {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else {
		creds, err := credentials.NewClientTLSFromFile(certificateFile, commonName)
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}

	err := protobuf.RegisterIndexHandlerFromEndpoint(ctx, mux, grpcAddress, dialOpts)
	if err != nil {
		logger.Error("failed to register KVS handler from endpoint", zap.Error(err))
		return nil, err
	}

	listener, err := net.Listen("tcp", httpAddress)
	if err != nil {
		logger.Error("failed to create index service", zap.Error(err))
		return nil, err
	}

	return &GRPCGateway{
		httpAddress:        httpAddress,
		grpcAddress:        grpcAddress,
		listener:           listener,
		mux:                mux,
		cancel:             cancel,
		certificateFile:    certificateFile,
		keyFile:            keyFile,
		corsAllowedMethods: corsAllowedMethods,
		corsAllowedOrigins: corsAllowedOrigins,
		corsAllowedHeaders: corsAllowedHeaders,
		logger:             logger,
	}, nil
}

func (s *GRPCGateway) Start() error {
	corsOpts := make([]handlers.CORSOption, 0)

	if s.corsAllowedMethods != nil && len(s.corsAllowedMethods) > 0 {
		corsOpts = append(corsOpts, handlers.AllowedMethods(s.corsAllowedMethods))
	}
	if s.corsAllowedOrigins != nil && len(s.corsAllowedOrigins) > 0 {
		corsOpts = append(corsOpts, handlers.AllowedMethods(s.corsAllowedOrigins))
	}
	if s.corsAllowedHeaders != nil && len(s.corsAllowedHeaders) > 0 {
		corsOpts = append(corsOpts, handlers.AllowedMethods(s.corsAllowedHeaders))
	}

	corsMux := handlers.CORS(
		corsOpts...,
	)(s.mux)

	if s.certificateFile == "" && s.keyFile == "" {
		go func() {
			if len(corsOpts) > 0 {
				_ = http.Serve(s.listener, corsMux)
			} else {
				_ = http.Serve(s.listener, s.mux)
			}
		}()
	} else {
		go func() {
			if len(corsOpts) > 0 {
				_ = http.ServeTLS(s.listener, corsMux, s.certificateFile, s.keyFile)
			} else {
				_ = http.ServeTLS(s.listener, s.mux, s.certificateFile, s.keyFile)
			}
		}()
	}

	s.logger.Info("gRPC gateway started", zap.String("http_address", s.httpAddress))
	return nil
}

func (s *GRPCGateway) Stop() error {
	defer s.cancel()

	err := s.listener.Close()
	if err != nil {
		s.logger.Error("failed to close listener", zap.String("http_address", s.listener.Addr().String()), zap.Error(err))
	}

	s.logger.Info("gRPC gateway stopped", zap.String("http_address", s.httpAddress))
	return nil
}
