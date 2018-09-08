//  Copyright (c) 2018 Minoru Osuka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mash/go-accesslog"
	"github.com/mosuka/blast/grpc/client"
	"github.com/mosuka/blast/http/handler"
	"github.com/mosuka/blast/logging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPServer struct {
	httpAddress      string
	listener         net.Listener
	httpAccessLogger *log.Logger

	grpcAddress string
	grpcClient  *client.GRPCClient

	logger *log.Logger
}

func NewHTTPServer(httpAddress string, grpcAddress string) (*HTTPServer, error) {
	return &HTTPServer{
		httpAddress:      httpAddress,
		grpcAddress:      grpcAddress,
		logger:           logging.DefaultLogger(),
		httpAccessLogger: logging.DefaultLogger(),
	}, nil
}

func (s *HTTPServer) SetLogger(logger *log.Logger) {
	s.logger = logger
	return
}

func (s *HTTPServer) SetHTTPAccessLogger(logger *log.Logger) {
	s.httpAccessLogger = logger
	return
}

func (s *HTTPServer) Start() error {
	var err error

	if s.grpcClient, err = client.NewGRPCClient(s.grpcAddress); err != nil {
		s.logger.Printf("[ERR] server: Failed to create gRPC client: %s", err.Error())
		return err
	}
	s.logger.Printf("[INFO] server: gRPC client has been created at %s", s.grpcAddress)

	if s.listener, err = net.Listen("tcp", s.httpAddress); err != nil {
		s.logger.Printf("[ERR] server: Failed to create listener: %s", err.Error())
		return err
	}
	s.logger.Printf("[INFO] server: Listener has been created at %s", s.httpAddress)

	router := mux.NewRouter()
	router.StrictSlash(true)

	// Set HTTP handlers
	router.Handle("/", handler.NewIndexHandler(s.logger)).Methods("GET")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	router.Handle("/rest/{id}", handler.NewPutHandler(s.logger, s.grpcClient)).Methods("PUT")
	router.Handle("/rest/{id}", handler.NewGetHandler(s.logger, s.grpcClient)).Methods("GET")
	router.Handle("/rest/{id}", handler.NewDeleteHandler(s.logger, s.grpcClient)).Methods("DELETE")
	router.Handle("/rest/_bulk", handler.NewBulkHandler(s.logger, s.grpcClient)).Methods("POST")
	router.Handle("/rest/_search", handler.NewSearchHandler(s.logger, s.grpcClient)).Methods("POST")

	go func() {
		// Start server
		s.logger.Print("[INFO] server: Start the HTTP server")
		http.Serve(
			s.listener,
			accesslog.NewLoggingHandler(
				router,
				logging.ApacheCombinedLogger{
					Logger: s.httpAccessLogger,
				},
			),
		)
		return
	}()

	return nil
}

func (s *HTTPServer) Stop() error {
	var err error

	if err = s.listener.Close(); err != nil {
		s.logger.Printf("[ERR] server: Failed to close listener: %s", err.Error())
	}
	s.logger.Print("[INFO] server: Listener has been closed")

	if err = s.grpcClient.Close(); err != nil {
		s.logger.Printf("[ERR] server: Failed to close gRPC client: %s", err.Error())
	}
	s.logger.Print("[INFO] server: gRPC client has been closed")

	s.logger.Print("[INFO] server: HTTP server has been stopped")
	return nil
}
