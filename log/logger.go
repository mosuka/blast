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

package log

import (
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/hashicorp/logutils"
	"github.com/mash/go-accesslog"
	"github.com/natefinch/lumberjack"
)

func NewFileWriter(filename string, maxSize int, maxBackups int, maxAge int, compress bool) io.Writer {
	var writer io.Writer
	if filename != "" {
		writer = &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    maxSize, // megabytes
			MaxBackups: maxBackups,
			MaxAge:     maxAge,   // days
			Compress:   compress, // disabled by default
		}
	} else {
		writer = os.Stderr
	}

	return writer
}

type callerInfo struct {
	packageName string
	fileName    string
	funcName    string
	line        int
}

func getCallerInfo() *callerInfo {
	pc, file, line, _ := runtime.Caller(5)
	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	packageName := ""
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		packageName = strings.Join(parts[0:pl-2], ".")
	} else {
		packageName = strings.Join(parts[0:pl-1], ".")
	}

	info := &callerInfo{
		packageName: packageName,
		fileName:    fileName,
		funcName:    funcName,
		line:        line,
	}

	return info
}

type CallerWriter struct {
	logger *log.Logger
	writer io.Writer
}

func NewCallerWriter(out io.Writer, prefix string, flag int) io.Writer {
	return &CallerWriter{
		logger: log.New(out, prefix, flag),
		writer: out,
	}
}

func (l *CallerWriter) Write(p []byte) (n int, err error) {
	info := getCallerInfo()
	l.logger.Printf("%s:%d %s", info.packageName+string(os.PathSeparator)+info.fileName, info.line, p)
	return len(p), nil
}

type LogLevel int

const (
	DEBUG = iota
	INFO
	WARN
	ERR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERR:
		return "ERR"
	default:
		return "INFO"
	}
}

func NewLogLevelFilter(logLevel string, writer io.Writer) io.Writer {
	filter := &logutils.LevelFilter{
		Levels: []logutils.LogLevel{
			logutils.LogLevel(LogLevel(DEBUG).String()),
			logutils.LogLevel(LogLevel(INFO).String()),
			logutils.LogLevel(LogLevel(WARN).String()),
			logutils.LogLevel(LogLevel(ERR).String()),
		},
		MinLevel: logutils.LogLevel(logLevel),
		Writer:   writer,
	}

	return filter
}

func DefaultLogger() *log.Logger {
	return Logger("DEBUG", "", log.LstdFlags|log.Lmicroseconds|log.LUTC, "", 0, 0, 0, false)
}

func Logger(logLevel string, prefix string, flag int, filename string, maxSize int, maxBackups int, maxAge int, compress bool) *log.Logger {
	fileWriter := NewFileWriter(filename, maxSize, maxBackups, maxAge, compress)

	callerWriter := NewCallerWriter(fileWriter, prefix, flag)

	logLevelFilter := NewLogLevelFilter(logLevel, callerWriter)

	logger := log.New(logLevelFilter, "", 0)

	return logger
}

func HTTPAccessLogger(filename string, maxSize int, maxBackups int, maxAge int, compress bool) *log.Logger {
	writer := NewFileWriter(filename, maxSize, maxBackups, maxAge, compress)

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
