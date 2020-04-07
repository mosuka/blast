package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/mosuka/blast/client"
	"github.com/mosuka/blast/log"
	"github.com/mosuka/blast/mapping"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the index server",
		Long:  "Start the index server",
		RunE: func(cmd *cobra.Command, args []string) error {
			id = viper.GetString("id")
			raftAddress = viper.GetString("raft_address")
			grpcAddress = viper.GetString("grpc_address")
			httpAddress = viper.GetString("http_address")
			dataDirectory = viper.GetString("data_directory")
			peerGrpcAddress = viper.GetString("peer_grpc_address")

			mappingFile = viper.GetString("mapping_file")

			certificateFile = viper.GetString("certificate_file")
			keyFile = viper.GetString("key_file")
			commonName = viper.GetString("common_name")

			logLevel = viper.GetString("log_level")
			logFile = viper.GetString("log_file")
			logMaxSize = viper.GetInt("log_max_size")
			logMaxBackups = viper.GetInt("log_max_backups")
			logMaxAge = viper.GetInt("log_max_age")
			logCompress = viper.GetBool("log_compress")

			logger := log.NewLogger(
				logLevel,
				logFile,
				logMaxSize,
				logMaxBackups,
				logMaxAge,
				logCompress,
			)

			bootstrap := peerGrpcAddress == "" || peerGrpcAddress == grpcAddress

			indexMapping := mapping.NewIndexMapping()
			if mappingFile != "" {
				var err error
				if indexMapping, err = mapping.NewIndexMappingFromFile(mappingFile); err != nil {
					return err
				}
			}

			raftServer, err := server.NewRaftServer(id, raftAddress, dataDirectory, indexMapping, bootstrap, logger)
			if err != nil {
				return err
			}

			grpcServer, err := server.NewGRPCServer(grpcAddress, raftServer, certificateFile, keyFile, commonName, logger)
			if err != nil {
				return err
			}

			grpcGateway, err := server.NewGRPCGateway(httpAddress, grpcAddress, certificateFile, keyFile, commonName, logger)
			if err != nil {
				return err
			}

			quitCh := make(chan os.Signal, 1)
			signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

			if err := raftServer.Start(); err != nil {
				return err
			}

			if err := grpcServer.Start(); err != nil {
				return err
			}

			if err := grpcGateway.Start(); err != nil {
				return err
			}

			// wait for detect leader if it's bootstrap
			if bootstrap {
				timeout := 60 * time.Second
				if err := raftServer.WaitForDetectLeader(timeout); err != nil {
					return err
				}
			}

			// create gRPC client for joining node
			var joinGrpcAddress string
			if bootstrap {
				joinGrpcAddress = grpcAddress
			} else {
				joinGrpcAddress = peerGrpcAddress
			}

			c, err := client.NewGRPCClientWithContextTLS(joinGrpcAddress, context.Background(), certificateFile, commonName)
			if err != nil {
				return err
			}
			defer func() {
				_ = c.Close()
			}()

			// join this node to the existing cluster
			joinRequest := &protobuf.JoinRequest{
				Id: id,
				Node: &protobuf.Node{
					RaftAddress: raftAddress,
					Metadata: &protobuf.Metadata{
						GrpcAddress: grpcAddress,
						HttpAddress: httpAddress,
					},
				},
			}
			if err = c.Join(joinRequest); err != nil {
				return err
			}

			// wait for receiving signal
			<-quitCh

			_ = grpcGateway.Stop()
			_ = grpcServer.Stop()
			_ = raftServer.Stop()

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(startCmd)

	cobra.OnInitialize(func() {
		if configFile != "" {
			viper.SetConfigFile(configFile)
		} else {
			home, err := homedir.Dir()
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			viper.AddConfigPath("/etc")
			viper.AddConfigPath(home)
			viper.SetConfigName("blast")
		}

		viper.SetEnvPrefix("BLAST")
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError:
				// config file does not found in search path
			default:
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	})

	startCmd.PersistentFlags().StringVar(&configFile, "config-file", "", "config file. if omitted, blast.yaml in /etc and home directory will be searched")
	startCmd.PersistentFlags().StringVar(&id, "id", "node1", "node ID")
	startCmd.PersistentFlags().StringVar(&raftAddress, "raft-address", ":7000", "Raft server listen address")
	startCmd.PersistentFlags().StringVar(&grpcAddress, "grpc-address", ":9000", "gRPC server listen address")
	startCmd.PersistentFlags().StringVar(&httpAddress, "http-address", ":8000", "HTTP server listen address")
	startCmd.PersistentFlags().StringVar(&dataDirectory, "data-directory", "/tmp/blast/data", "data directory which store the index and Raft logs")
	startCmd.PersistentFlags().StringVar(&peerGrpcAddress, "peer-grpc-address", "", "listen address of the existing gRPC server in the joining cluster")
	startCmd.PersistentFlags().StringVar(&mappingFile, "mapping-file", "", "path to the index mapping file")
	startCmd.PersistentFlags().StringVar(&certificateFile, "certificate-file", "", "path to the client server TLS certificate file")
	startCmd.PersistentFlags().StringVar(&keyFile, "key-file", "", "path to the client server TLS key file")
	startCmd.PersistentFlags().StringVar(&commonName, "common-name", "", "certificate common name")
	startCmd.PersistentFlags().StringVar(&logLevel, "log-level", "INFO", "log level")
	startCmd.PersistentFlags().StringVar(&logFile, "log-file", os.Stderr.Name(), "log file")
	startCmd.PersistentFlags().IntVar(&logMaxSize, "log-max-size", 500, "max size of a log file in megabytes")
	startCmd.PersistentFlags().IntVar(&logMaxBackups, "log-max-backups", 3, "max backup count of log files")
	startCmd.PersistentFlags().IntVar(&logMaxAge, "log-max-age", 30, "max age of a log file in days")
	startCmd.PersistentFlags().BoolVar(&logCompress, "log-compress", false, "compress a log file")

	_ = viper.BindPFlag("id", startCmd.PersistentFlags().Lookup("id"))
	_ = viper.BindPFlag("raft_address", startCmd.PersistentFlags().Lookup("raft-address"))
	_ = viper.BindPFlag("grpc_address", startCmd.PersistentFlags().Lookup("grpc-address"))
	_ = viper.BindPFlag("http_address", startCmd.PersistentFlags().Lookup("http-address"))
	_ = viper.BindPFlag("data_directory", startCmd.PersistentFlags().Lookup("data-directory"))
	_ = viper.BindPFlag("peer_grpc_address", startCmd.PersistentFlags().Lookup("peer-grpc-address"))
	_ = viper.BindPFlag("mapping_file", startCmd.PersistentFlags().Lookup("mapping-file"))
	_ = viper.BindPFlag("certificate_file", startCmd.PersistentFlags().Lookup("certificate-file"))
	_ = viper.BindPFlag("key_file", startCmd.PersistentFlags().Lookup("key-file"))
	_ = viper.BindPFlag("common_name", startCmd.PersistentFlags().Lookup("common-name"))
	_ = viper.BindPFlag("log_level", startCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("log_max_size", startCmd.PersistentFlags().Lookup("log-max-size"))
	_ = viper.BindPFlag("log_max_backups", startCmd.PersistentFlags().Lookup("log-max-backups"))
	_ = viper.BindPFlag("log_max_age", startCmd.PersistentFlags().Lookup("log-max-age"))
	_ = viper.BindPFlag("log_compress", startCmd.PersistentFlags().Lookup("log-compress"))
}
