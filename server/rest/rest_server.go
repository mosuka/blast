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

package rest

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mosuka/blast/client"
	"github.com/mosuka/blast/server/rest/handler"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
)

type blastRESTServer struct {
	router   *mux.Router
	listener net.Listener
	client   *client.BlastClientWrapper
}

func NewBlastRESTServer(port int, basePath, server string) *blastRESTServer {
	router := mux.NewRouter()
	router.StrictSlash(true)

	c, err := client.NewBlastClientWrapper(server)
	if err != nil {
		return nil
	}

	/*
	 * set handlers
	 */
	router.Handle(fmt.Sprintf("/%s/", basePath), handler.NewGetIndexHandler(c)).Methods("GET")
	router.Handle(fmt.Sprintf("/%s/{id}", basePath), handler.NewPutDocumentHandler(c)).Methods("PUT")
	router.Handle(fmt.Sprintf("/%s/{id}", basePath), handler.NewGetDocumentHandler(c)).Methods("GET")
	router.Handle(fmt.Sprintf("/%s/{id}", basePath), handler.NewDeleteDocumentHandler(c)).Methods("DELETE")
	router.Handle(fmt.Sprintf("/%s/_bulk", basePath), handler.NewBulkHandler(c)).Methods("POST")
	router.Handle(fmt.Sprintf("/%s/_search", basePath), handler.NewSearchHandler(c)).Methods("POST")

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
		client:   c,
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
