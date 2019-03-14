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
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace = "blast"
	subsystem = "http"

	DurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "duration_seconds",
			Help:      "The invocation duration in seconds.",
		},
		[]string{
			"request_uri",
		},
	)

	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "The number of requests.",
		},
		[]string{
			"request_uri",
			"method",
		},
	)

	ResponsesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "responses_total",
			Help:      "The number of responses.",
		},
		[]string{
			"request_uri",
			"status",
		},
	)

	RequestsBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_bytes_total",
			Help:      "A summary of the invocation requests bytes.",
		},
		[]string{
			"request_uri",
			"method",
		},
	)

	ResponsesBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "responses_bytes_total",
			Help:      "A summary of the invocation responses bytes.",
		},
		[]string{
			"request_uri",
			"method",
		},
	)
)

func init() {
	prometheus.MustRegister(DurationSeconds)
	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(ResponsesTotal)
	prometheus.MustRegister(RequestsBytesTotal)
	prometheus.MustRegister(ResponsesBytesTotal)
}

func RecordMetrics(start time.Time, status int, writer http.ResponseWriter, request *http.Request, logger *log.Logger) {
	DurationSeconds.With(prometheus.Labels{"request_uri": request.RequestURI}).Observe(float64(time.Since(start)) / float64(time.Second))
	RequestsTotal.With(prometheus.Labels{"request_uri": request.RequestURI, "method": request.Method}).Inc()
	ResponsesTotal.With(prometheus.Labels{"request_uri": request.RequestURI, "status": strconv.Itoa(status)}).Inc()
	RequestsBytesTotal.With(prometheus.Labels{"request_uri": request.RequestURI, "method": request.Method}).Add(float64(request.ContentLength))
	contentLength, err := strconv.ParseFloat(writer.Header().Get("Content-Length"), 64)
	if err != nil {
		logger.Printf("[ERR] Failed to parse content length: %v", err)
	}
	ResponsesBytesTotal.With(prometheus.Labels{"request_uri": request.RequestURI, "method": request.Method}).Add(contentLength)
}
