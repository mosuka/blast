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

package cmd

import (
	"fmt"
	"github.com/mosuka/blast/server"
	"github.com/mosuka/blast/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type RootCommandOptions struct {
	config         string
	logFormat      string
	logOutput      string
	logLevel       string
	port           int
	server         string
	baseURI        string
	dialTimeout    int
	requestTimeout int
	versionFlag    bool
}

var rootCmdOpts = RootCommandOptions{
	config:         "",
	logFormat:      "text",
	logOutput:      "",
	logLevel:       "info",
	port:           8080,
	server:         "localhost:5000",
	baseURI:        "api",
	dialTimeout:    15000,
	requestTimeout: 15000,
	versionFlag:    false,
}

var logOutput *os.File

var RootCmd = &cobra.Command{
	Use:                "blastrest",
	Short:              "Blast REST server",
	Long:               `The Command Line Interface for the Blast REST server.`,
	PersistentPreRunE:  persistentPreRunERootCmd,
	RunE:               runERootCmd,
	PersistentPostRunE: persistentPostRunERootCmd,
}

func persistentPreRunERootCmd(cmd *cobra.Command, args []string) error {
	if rootCmdOpts.versionFlag {
		fmt.Printf("%s\n", version.Version)
		os.Exit(0)
	}

	switch viper.GetString("log_format") {
	case "text":
		log.SetFormatter(&log.TextFormatter{
			ForceColors:      false,
			DisableColors:    true,
			DisableTimestamp: false,
			FullTimestamp:    true,
			TimestampFormat:  time.RFC3339,
			DisableSorting:   false,
			QuoteEmptyFields: true,
			QuoteCharacter:   "\"",
		})
	case "color":
		log.SetFormatter(&log.TextFormatter{
			ForceColors:      true,
			DisableColors:    false,
			DisableTimestamp: false,
			FullTimestamp:    true,
			TimestampFormat:  time.RFC3339,
			DisableSorting:   false,
			QuoteEmptyFields: true,
			QuoteCharacter:   "\"",
		})
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat:  time.RFC3339,
			DisableTimestamp: false,
			FieldMap: log.FieldMap{
				log.FieldKeyTime:  "@timestamp",
				log.FieldKeyLevel: "@level",
				log.FieldKeyMsg:   "@message",
			},
		})
	default:
		log.SetFormatter(&log.TextFormatter{
			ForceColors:      false,
			DisableColors:    true,
			DisableTimestamp: false,
			FullTimestamp:    true,
			TimestampFormat:  time.RFC3339,
			DisableSorting:   false,
			QuoteEmptyFields: true,
			QuoteCharacter:   "\"",
		})
	}

	switch viper.GetString("log_level") {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	if viper.GetString("log_output") == "" {
		log.SetOutput(os.Stdout)
	} else {
		var err error
		logOutput, err = os.OpenFile(viper.GetString("log_output"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		} else {
			log.SetOutput(logOutput)
		}
	}

	return nil
}

func runERootCmd(cmd *cobra.Command, args []string) error {
	s := server.NewBlastRESTServer(viper.GetInt("port"), viper.GetString("base_uri"), viper.GetString("server"), viper.GetInt("dial_timeout"), viper.GetInt("request_timeout"))
	s.Start()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	for {
		sig := <-signalChan

		log.WithFields(log.Fields{
			"signal": sig,
		}).Info("trap signal")

		s.Stop()

		return nil
	}

	return nil
}

func persistentPostRunERootCmd(cmd *cobra.Command, args []string) error {
	if viper.GetString("log_output") != "" {
		logOutput.Close()
	}

	return nil
}

func LoadConfig() {
	viper.SetDefault("log_format", rootCmdOpts.logFormat)
	viper.SetDefault("log_output", rootCmdOpts.logOutput)
	viper.SetDefault("log_level", rootCmdOpts.logLevel)

	viper.SetDefault("port", rootCmdOpts.port)
	viper.SetDefault("base_uri", rootCmdOpts.baseURI)
	viper.SetDefault("server", rootCmdOpts.server)
	viper.SetDefault("timeout", rootCmdOpts.requestTimeout)

	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
	} else {
		viper.SetConfigName("blastrest")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("/etc")
		viper.AddConfigPath("${HOME}/etc")
		viper.AddConfigPath("./etc")
	}
	viper.SetEnvPrefix("blastrest")
	viper.AutomaticEnv()

	viper.ReadInConfig()
}

func init() {
	cobra.OnInitialize(LoadConfig)

	RootCmd.Flags().SortFlags = false

	RootCmd.Flags().String("config", rootCmdOpts.config, "config file path")
	RootCmd.Flags().String("log-format", rootCmdOpts.logFormat, "log format")
	RootCmd.Flags().String("log-output", rootCmdOpts.logOutput, "log output path")
	RootCmd.Flags().String("log-level", rootCmdOpts.logLevel, "log level")
	RootCmd.Flags().Int("port", rootCmdOpts.port, "port number")
	RootCmd.Flags().String("base-uri", rootCmdOpts.baseURI, "base URI of API endpoint")
	RootCmd.Flags().String("server", rootCmdOpts.server, "server to connect to")
	RootCmd.Flags().Int("dial-timeout", rootCmdOpts.dialTimeout, "dial timeout")
	RootCmd.Flags().Int("request-timeout", rootCmdOpts.requestTimeout, "request timeout")
	RootCmd.Flags().BoolVarP(&rootCmdOpts.versionFlag, "version", "v", rootCmdOpts.versionFlag, "show version numner")

	viper.BindPFlag("config", RootCmd.Flags().Lookup("config"))
	viper.BindPFlag("log_format", RootCmd.Flags().Lookup("log-format"))
	viper.BindPFlag("log_output", RootCmd.Flags().Lookup("log-output"))
	viper.BindPFlag("log_level", RootCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("port", RootCmd.Flags().Lookup("port"))
	viper.BindPFlag("base_uri", RootCmd.Flags().Lookup("base-uri"))
	viper.BindPFlag("server", RootCmd.Flags().Lookup("server"))
	viper.BindPFlag("dial_timeout", RootCmd.Flags().Lookup("dial-timeout"))
	viper.BindPFlag("request_timeout", RootCmd.Flags().Lookup("request-timeout"))
}
