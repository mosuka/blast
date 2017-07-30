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
	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/server"
	"github.com/mosuka/blast/util"
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
	etcdServers    []string
	requestTimeout int
	clusterName    string
	port           int
	indexPath      string
	indexMapping   string
	indexType      string
	kvstore        string
	kvconfig       string
	versionFlag    bool
}

var rootCmdOpts = RootCommandOptions{
	config:         "",
	logFormat:      "text",
	logOutput:      "",
	logLevel:       "info",
	port:           20884,
	etcdServers:    []string{},
	requestTimeout: 15000,
	clusterName:    "",
	indexPath:      "./data/index",
	indexMapping:   "",
	indexType:      "upside_down",
	kvstore:        "boltdb",
	kvconfig:       "",
	versionFlag:    false,
}

var logOutput *os.File

var RootCmd = &cobra.Command{
	Use:                "blast",
	Short:              "Blast",
	Long:               `The Command Line Interface for the Blast.`,
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
	// server
	svr := server.NewBlastServer()

	if len(viper.GetStringSlice("etcd_servers")) > 0 && viper.GetString("cluster_name") != "" {
		err := svr.ConnectEtcd(viper.GetStringSlice("etcd_servers"), viper.GetInt("request_timeout"))
		if err != nil {
			return err
		}

		indexMapping, err := svr.GetIndexMappingFromEtc(viper.GetString("cluster_name"))
		if err != nil {
			return err
		}

		indexType, err := svr.GetIndexTypeFromEtc(viper.GetString("cluster_name"))
		if err != nil {
			return err
		}

		kvstore, err := svr.GetKvstoreFromEtc(viper.GetString("cluster_name"))
		if err != nil {
			return err
		}

		kvconfig, err := svr.GetKvconfigFromEtc(viper.GetString("cluster_name"))
		if err != nil {
			return err
		}

		svr.Start(viper.GetInt("port"), viper.GetString("index_path"), indexMapping, indexType, kvstore, kvconfig)

		svr.JoinCluster(viper.GetString("cluster_name"))
	} else {
		var indexMapping *mapping.IndexMappingImpl
		if viper.GetString("index_mapping") != "" {
			file, err := os.Open(viper.GetString("index_mapping"))
			if err != nil {
				return err
			}
			defer file.Close()

			indexMapping, err = util.NewIndexMapping(file)
			if err != nil {
				return err
			}
		}

		var kvconfig map[string]interface{}
		if viper.GetString("kvconfig") != "" {
			file, err := os.Open(viper.GetString("kvconfig"))
			if err != nil {
				return err
			}
			defer file.Close()

			kvconfig, err = util.NewKvconfig(file)
			if err != nil {
				return err
			}
		}

		svr.Start(viper.GetInt("port"), viper.GetString("index_path"), indexMapping, viper.GetString("index_type"), viper.GetString("kvstore"), kvconfig)
	}

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

		if viper.GetString("cluster_name") != "" {
			svr.LeaveCluster(viper.GetString("cluster_name"))
		}

		svr.Stop()

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
	viper.SetDefault("etcd_endpoints", rootCmdOpts.etcdServers)
	viper.SetDefault("request_timeout", rootCmdOpts.requestTimeout)
	viper.SetDefault("cluster_name", rootCmdOpts.clusterName)
	viper.SetDefault("index_path", rootCmdOpts.indexPath)
	viper.SetDefault("index_mapping", rootCmdOpts.indexMapping)
	viper.SetDefault("index_type", rootCmdOpts.indexType)
	viper.SetDefault("kvstore", rootCmdOpts.kvstore)
	viper.SetDefault("kvconfig", rootCmdOpts.kvconfig)

	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
	} else {
		viper.SetConfigName("blast")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("/etc")
		viper.AddConfigPath("${HOME}/etc")
		viper.AddConfigPath("./etc")
	}
	viper.SetEnvPrefix("blast")
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
	RootCmd.Flags().StringSlice("etcd-server", rootCmdOpts.etcdServers, "etcd server")
	RootCmd.Flags().Int("request-timeout", rootCmdOpts.requestTimeout, "request timeout")
	RootCmd.Flags().String("cluster-name", rootCmdOpts.clusterName, "cluster name")
	RootCmd.Flags().String("index-path", rootCmdOpts.indexPath, "index directory path")
	RootCmd.Flags().String("index-mapping", rootCmdOpts.indexMapping, "index mapping path")
	RootCmd.Flags().String("index-type", rootCmdOpts.indexType, "index type")
	RootCmd.Flags().String("kvstore", rootCmdOpts.kvstore, "kvstore")
	RootCmd.Flags().String("kvconfig", rootCmdOpts.kvconfig, "kvconfig path")
	RootCmd.Flags().BoolVarP(&rootCmdOpts.versionFlag, "version", "v", rootCmdOpts.versionFlag, "show version number")

	viper.BindPFlag("config", RootCmd.Flags().Lookup("config"))
	viper.BindPFlag("log_format", RootCmd.Flags().Lookup("log-format"))
	viper.BindPFlag("log_output", RootCmd.Flags().Lookup("log-output"))
	viper.BindPFlag("log_level", RootCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("port", RootCmd.Flags().Lookup("port"))
	viper.BindPFlag("etcd_servers", RootCmd.Flags().Lookup("etcd-server"))
	viper.BindPFlag("request_timeout", RootCmd.Flags().Lookup("request-timeout"))
	viper.BindPFlag("cluster_name", RootCmd.Flags().Lookup("cluster-name"))
	viper.BindPFlag("index_path", RootCmd.Flags().Lookup("index-path"))
	viper.BindPFlag("index_mapping", RootCmd.Flags().Lookup("index-mapping"))
	viper.BindPFlag("index_type", RootCmd.Flags().Lookup("index-type"))
	viper.BindPFlag("kvstore", RootCmd.Flags().Lookup("kvstore"))
	viper.BindPFlag("kvconfig", RootCmd.Flags().Lookup("kvconfig"))
}
