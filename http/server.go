// Copyright (c) 2019 Minoru Osuka
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

package http

import (
	"net"
	"net/http"

	accesslog "github.com/mash/go-accesslog"
	"go.uber.org/zap"
)

type Server struct {
	listener net.Listener
	router   *Router

	logger     *zap.Logger
	httpLogger accesslog.Logger
}

func NewServer(httpAddr string, router *Router, logger *zap.Logger, httpLogger accesslog.Logger) (*Server, error) {
	listener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		return nil, err
	}

	return &Server{
		listener:   listener,
		router:     router,
		logger:     logger,
		httpLogger: httpLogger,
	}, nil
}

func (s *Server) Start() error {
	err := http.Serve(
		s.listener,
		accesslog.NewLoggingHandler(
			s.router,
			s.httpLogger,
		),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	err := s.listener.Close()
	if err != nil {
		return err
	}

	return nil
}
