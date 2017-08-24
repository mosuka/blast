//  Copyright (c) 2017 Minoru Osuka
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
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mosuka/blast/client"
	"github.com/mosuka/blast/server/handler"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

type blastRESTServer struct {
	router      *mux.Router
	listener    net.Listener
	client      *client.Client
	dialTimeout int
}

func NewBlastRESTServer(port int, basePath, server string, dialTimeout int, requestTimeout int) *blastRESTServer {
	router := mux.NewRouter()
	router.StrictSlash(true)

	// create client config
	cfg := client.Config{
		Server:      server,
		DialTimeout: time.Duration(dialTimeout) * time.Millisecond,
	}

	// create client
	clt, err := client.NewClient(&cfg)
	if err != nil {
		return nil
	}

	/*
	 * set handlers
	 */
	router.Handle(fmt.Sprintf("/%s/", basePath), handler.NewGetIndexInfoHandler(clt)).Methods("GET")
	router.Handle(fmt.Sprintf("/%s/{id}", basePath), handler.NewPutDocumentHandler(clt)).Methods("PUT")
	router.Handle(fmt.Sprintf("/%s/{id}", basePath), handler.NewGetDocumentHandler(clt)).Methods("GET")
	router.Handle(fmt.Sprintf("/%s/{id}", basePath), handler.NewDeleteDocumentHandler(clt)).Methods("DELETE")
	router.Handle(fmt.Sprintf("/%s/_bulk", basePath), handler.NewBulkHandler(clt)).Methods("POST")
	router.Handle(fmt.Sprintf("/%s/_search", basePath), handler.NewSearchHandler(clt)).Methods("POST")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err == nil {
		log.WithFields(log.Fields{
			"port": port,
		}).Info("succeeded in creating listener")
	} else {
		log.WithFields(log.Fields{
			"port": port,
			"err":  err,
		}).Error("failed to create listener")
		return nil
	}

	return &blastRESTServer{
		router:   router,
		listener: listener,
		client:   clt,
	}
}

func (s *blastRESTServer) Start() error {
	go func() {
		http.Serve(s.listener, s.router)
		return
	}()

	log.WithFields(log.Fields{
		"addr": s.listener.Addr().String(),
	}).Info("The Blast REST Server started")

	return nil
}

func (s *blastRESTServer) Stop() error {
	err := s.client.Close()
	if err == nil {
		log.Info("succeeded in closing connection")
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to close connection")

		return err
	}

	err = s.listener.Close()
	if err == nil {
		log.Info("succeeded in closing listener")
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to close listener")

		return err
	}

	log.WithFields(log.Fields{
		"addr": s.listener.Addr().String(),
	}).Info("The Blast REST Server stopped")

	return nil
}
