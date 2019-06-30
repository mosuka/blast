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
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace = "http"
	subsystem = "server"

	DurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "handling_seconds",
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
			Name:      "handled_total",
			Help:      "The number of requests.",
		},
		[]string{
			"request_uri",
			"http_method",
			"http_status",
		},
	)

	RequestsBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_received_bytes",
			Help:      "A summary of the invocation requests bytes.",
		},
		[]string{
			"request_uri",
			"http_method",
		},
	)

	ResponsesBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "responses_sent_bytes",
			Help:      "A summary of the invocation responses bytes.",
		},
		[]string{
			"request_uri",
			"http_method",
		},
	)
)

func init() {
	prometheus.MustRegister(DurationSeconds)
	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(RequestsBytesTotal)
	prometheus.MustRegister(ResponsesBytesTotal)
}

func RecordMetrics(start time.Time, status int, writer http.ResponseWriter, request *http.Request) {
	DurationSeconds.With(prometheus.Labels{"request_uri": request.RequestURI}).Observe(float64(time.Since(start)) / float64(time.Second))

	RequestsTotal.With(prometheus.Labels{"request_uri": request.RequestURI, "http_method": request.Method, "http_status": strconv.Itoa(status)}).Inc()

	RequestsBytesTotal.With(prometheus.Labels{"request_uri": request.RequestURI, "http_method": request.Method}).Add(float64(request.ContentLength))

	contentLength, err := strconv.ParseFloat(writer.Header().Get("Content-Length"), 64)
	if err == nil {
		ResponsesBytesTotal.With(prometheus.Labels{"request_uri": request.RequestURI, "http_method": request.Method}).Add(contentLength)
	}
}
