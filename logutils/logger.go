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

package logutils

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(logLevel string, logFilename string, logMaxSize int, logMaxBackups int, logMaxAge int, logCompress bool) *zap.Logger {
	var ll zapcore.Level
	switch logLevel {
	case "DEBUG":
		ll = zap.DebugLevel
	case "INFO":
		ll = zap.InfoLevel
	case "WARN", "WARNING":
		ll = zap.WarnLevel
	case "ERR", "ERROR":
		ll = zap.WarnLevel
	case "DPANIC":
		ll = zap.DPanicLevel
	case "PANIC":
		ll = zap.PanicLevel
	case "FATAL":
		ll = zap.FatalLevel
	}

	var ws zapcore.WriteSyncer
	if logFilename == "" {
		ws = zapcore.AddSync(os.Stderr)
	} else {
		ws = zapcore.AddSync(
			&lumberjack.Logger{
				Filename:   logFilename,
				MaxSize:    logMaxSize, // megabytes
				MaxBackups: logMaxBackups,
				MaxAge:     logMaxAge, // days
				Compress:   logCompress,
			},
		)
	}

	ec := zap.NewProductionEncoderConfig()
	ec.TimeKey = "_timestamp_"
	ec.LevelKey = "_level_"
	ec.NameKey = "_name_"
	ec.CallerKey = "_caller_"
	ec.MessageKey = "_message_"
	ec.StacktraceKey = "_stacktrace_"
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	ec.EncodeCaller = zapcore.ShortCallerEncoder

	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(ec),
			ws,
			ll,
		),
		zap.AddCaller(),
		//zap.AddStacktrace(ll),
	)

	return logger
}
