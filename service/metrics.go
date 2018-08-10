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

package service

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace = "blast"
	subsystem = "service"

	DurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "duration_seconds",
			Help:      "The gRPC invocation durations in seconds.",
		},
		[]string{
			"func",
		},
	)

	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "The number of gRPC requests.",
		},
		[]string{
			"func",
		},
	)
)

func init() {
	prometheus.MustRegister(DurationSeconds)
	prometheus.MustRegister(RequestsTotal)
}

func Metrics(start time.Time, funcName string) error {
	DurationSeconds.With(prometheus.Labels{"func": funcName}).Observe(float64(time.Since(start)) / float64(time.Second))
	RequestsTotal.With(prometheus.Labels{"func": funcName}).Inc()

	return nil
}
