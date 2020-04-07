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
	leaveCmd = &cobra.Command{
		Use:   "leave ID",
		Args:  cobra.ExactArgs(1),
		Short: "Leave a node from the cluster",
		Long:  "Leave a node from the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			grpcAddress = viper.GetString("grpc_address")

			certificateFile = viper.GetString("certificate_file")
			commonName = viper.GetString("common_name")

			id := args[0]

			c, err := client.NewGRPCClientWithContextTLS(grpcAddress, context.Background(), certificateFile, commonName)
			if err != nil {
				return err
			}
			defer func() {
				_ = c.Close()
			}()

			req := &protobuf.LeaveRequest{
				Id: id,
			}

			if err := c.Leave(req); err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(leaveCmd)

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
				// config file does not found in config search path
			default:
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	})

	leaveCmd.PersistentFlags().StringVar(&configFile, "config-file", "", "config file. if omitted, blast.yaml in /etc and home directory will be searched")
	leaveCmd.PersistentFlags().StringVar(&grpcAddress, "grpc-address", ":9000", "gRPC server listen address")
	leaveCmd.PersistentFlags().StringVar(&certificateFile, "certificate-file", "", "path to the client server TLS certificate file")
	leaveCmd.PersistentFlags().StringVar(&commonName, "common-name", "", "certificate common name")

	_ = viper.BindPFlag("grpc_address", leaveCmd.PersistentFlags().Lookup("grpc-address"))
	_ = viper.BindPFlag("certificate_file", leaveCmd.PersistentFlags().Lookup("certificate-file"))
	_ = viper.BindPFlag("common_name", leaveCmd.PersistentFlags().Lookup("common-name"))
}
