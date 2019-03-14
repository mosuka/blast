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
	"strconv"

	accesslog "github.com/mash/go-accesslog"
	blastlog "github.com/mosuka/blast/log"
)

func NewLogger(filename string, maxSize int, maxBackups int, maxAge int, compress bool) *log.Logger {
	writer := blastlog.NewFileWriter(filename, maxSize, maxBackups, maxAge, compress)

	logger := log.New(writer, "", 0)

	return logger
}

type ApacheCombinedLogger struct {
	Logger *log.Logger
}

func (l ApacheCombinedLogger) Log(record accesslog.LogRecord) {
	// Output log that formatted Apache combined.
	size := "-"
	if record.Size > 0 {
		size = strconv.FormatInt(record.Size, 10)
	}

	referer := "-"
	if record.RequestHeader.Get("Referer") != "" {
		referer = record.RequestHeader.Get("Referer")
	}

	userAgent := "-"
	if record.RequestHeader.Get("User-Agent") != "" {
		userAgent = record.RequestHeader.Get("User-Agent")
	}

	l.Logger.Printf(
		"%s - %s [%s] \"%s %s %s\" %d %s \"%s\" \"%s\"",
		record.Ip,
		record.Username,
		record.Time.Format("02/Jan/2006 03:04:05 +0000"),
		record.Method,
		record.Uri,
		record.Protocol,
		record.Status,
		size,
		referer,
		userAgent,
	)
}
