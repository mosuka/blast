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
	"github.com/mosuka/blast/server/grpc"
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
	versionFlag bool
}

var rootCmdOpts RootCommandOptions

var logOutput *os.File

var RootCmd = &cobra.Command{
	Use:                "blast",
	Short:              "Blast Server",
	Long:               `The Command Line Interface for the Blast Server.`,
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
	// IndexMapping
	indexMapping := mapping.NewIndexMapping()
	if cmd.Flag("index-mapping").Changed {
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

	// Kvconfig
	kvconfig := make(map[string]interface{})
	if cmd.Flag("kvconfig").Changed {
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

	server := grpc.NewBlastGRPCServer(viper.GetInt("port"), viper.GetString("index_path"), indexMapping, viper.GetString("index_type"), viper.GetString("kvstore"), kvconfig)
	server.Start(viper.GetBool("delete_index_at_startup"))

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

		server.Stop(viper.GetBool("delete_index_at_shutdown"))

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
	viper.SetDefault("log_format", DefaultLogFormat)
	viper.SetDefault("log_output", DefaultLogOutput)
	viper.SetDefault("log_level", DefaultLogLevel)
	viper.SetDefault("port", DefaultPort)
	viper.SetDefault("index_path", DefaultIndexPath)
	viper.SetDefault("index_mapping", DefaultIndexMapping)
	viper.SetDefault("index_type", DefaultIndexType)
	viper.SetDefault("kvstore", DefaultKvstore)
	viper.SetDefault("kvconfig", DefaultKvconfig)
	viper.SetDefault("delete_index_at_startup", DefaultDeleteIndexAtStartup)
	viper.SetDefault("delete_index_at_shutdown", DefaultDeleteIndexAtShutdown)

	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
	} else {
		viper.SetConfigName("blast")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("/etc/blast")
		viper.AddConfigPath("${HOME}/blast")
		viper.AddConfigPath("./blast")
	}
	viper.SetEnvPrefix("blast")
	viper.AutomaticEnv()

	viper.ReadInConfig()
}

func init() {
	cobra.OnInitialize(LoadConfig)

	RootCmd.PersistentFlags().BoolVar(&rootCmdOpts.versionFlag, "version", false, "show version number")
	RootCmd.PersistentFlags().String("config", DefaultConfig, "config file path")
	RootCmd.Flags().String("log-format", DefaultLogFormat, "log format")
	RootCmd.Flags().String("log-output", DefaultLogOutput, "log output path")
	RootCmd.Flags().String("log-level", DefaultLogLevel, "log level")
	RootCmd.Flags().Int("port", DefaultPort, "port number")
	RootCmd.Flags().String("index-path", DefaultIndexPath, "index directory path")
	RootCmd.Flags().String("index-mapping", DefaultIndexMapping, "index mapping path")
	RootCmd.Flags().String("index-type", DefaultIndexType, "index type")
	RootCmd.Flags().String("kvstore", DefaultKvstore, "kvstore")
	RootCmd.Flags().String("kvconfig", DefaultKvconfig, "kvconfig path")
	RootCmd.Flags().Bool("delete-index-at-startup", DefaultDeleteIndexAtStartup, "delete index at startup")
	RootCmd.Flags().Bool("delete-index-at-shutdown", DefaultDeleteIndexAtShutdown, "delete index at shutdown")

	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("log_format", RootCmd.Flags().Lookup("log-format"))
	viper.BindPFlag("log_output", RootCmd.Flags().Lookup("log-output"))
	viper.BindPFlag("log_level", RootCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("port", RootCmd.Flags().Lookup("port"))
	viper.BindPFlag("index_path", RootCmd.Flags().Lookup("index-path"))
	viper.BindPFlag("index_mapping", RootCmd.Flags().Lookup("index-mapping"))
	viper.BindPFlag("index_type", RootCmd.Flags().Lookup("index-type"))
	viper.BindPFlag("kvstore", RootCmd.Flags().Lookup("kvstore"))
	viper.BindPFlag("kvconfig", RootCmd.Flags().Lookup("kvconfig"))
	viper.BindPFlag("delete_index_at_startup", RootCmd.Flags().Lookup("delete-index-at-startup"))
	viper.BindPFlag("delete_index_at_shutdown", RootCmd.Flags().Lookup("delete-index-at-shutdown"))
}
