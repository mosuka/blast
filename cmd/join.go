package cmd

import (
	"context"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/mosuka/blast/client"
	"github.com/mosuka/blast/protobuf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	joinCmd = &cobra.Command{
		Use:   "join ID GRPC_ADDRESS",
		Args:  cobra.ExactArgs(2),
		Short: "Join a node to the cluster",
		Long:  "Join a node to the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			grpcAddress = viper.GetString("grpc_address")

			certificateFile = viper.GetString("certificate_file")
			commonName = viper.GetString("common_name")

			id := args[0]
			targetGrpcAddress := args[1]

			t, err := client.NewGRPCClientWithContextTLS(targetGrpcAddress, context.Background(), certificateFile, commonName)
			if err != nil {
				return err
			}
			defer func() {
				_ = t.Close()
			}()

			nodeResp, err := t.Node()
			if err != nil {
				return err
			}

			c, err := client.NewGRPCClientWithContextTLS(grpcAddress, context.Background(), certificateFile, commonName)
			if err != nil {
				return err
			}
			defer func() {
				_ = c.Close()
			}()

			req := &protobuf.JoinRequest{
				Id:   id,
				Node: nodeResp.Node,
			}

			if err := c.Join(req); err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(joinCmd)

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

	joinCmd.PersistentFlags().StringVar(&configFile, "config-file", "", "config file. if omitted, blast.yaml in /etc and home directory will be searched")
	joinCmd.PersistentFlags().StringVar(&grpcAddress, "grpc-address", ":9000", "gRPC server listen address")
	joinCmd.PersistentFlags().StringVar(&certificateFile, "certificate-file", "", "path to the client server TLS certificate file")
	joinCmd.PersistentFlags().StringVar(&commonName, "common-name", "", "certificate common name")

	_ = viper.BindPFlag("grpc_address", joinCmd.PersistentFlags().Lookup("grpc-address"))
	_ = viper.BindPFlag("certificate_file", joinCmd.PersistentFlags().Lookup("certificate-file"))
	_ = viper.BindPFlag("common_name", joinCmd.PersistentFlags().Lookup("common-name"))
}
