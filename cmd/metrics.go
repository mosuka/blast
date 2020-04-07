package cmd

import (
	"context"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/mosuka/blast/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	metricsCmd = &cobra.Command{
		Use:   "metrics",
		Short: "Get the node metrics",
		Long:  "Get the node metrics in Prometheus exposition format",
		RunE: func(cmd *cobra.Command, args []string) error {
			grpcAddress = viper.GetString("grpc_address")

			certificateFile = viper.GetString("certificate_file")
			commonName = viper.GetString("common_name")

			c, err := client.NewGRPCClientWithContextTLS(grpcAddress, context.Background(), certificateFile, commonName)
			if err != nil {
				return err
			}
			defer func() {
				_ = c.Close()
			}()

			resp, err := c.Metrics()
			if err != nil {
				return err
			}

			fmt.Println(string(resp.Metrics))

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(metricsCmd)

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

	metricsCmd.PersistentFlags().StringVar(&configFile, "config-file", "", "config file. if omitted, blast.yaml in /etc and home directory will be searched")
	metricsCmd.PersistentFlags().StringVar(&grpcAddress, "grpc-address", ":9000", "gRPC server listen address")
	metricsCmd.PersistentFlags().StringVar(&certificateFile, "certificate-file", "", "path to the client server TLS certificate file")
	metricsCmd.PersistentFlags().StringVar(&commonName, "common-name", "", "certificate common name")

	_ = viper.BindPFlag("grpc_address", metricsCmd.PersistentFlags().Lookup("grpc-address"))
	_ = viper.BindPFlag("certificate_file", metricsCmd.PersistentFlags().Lookup("certificate-file"))
	_ = viper.BindPFlag("common_name", metricsCmd.PersistentFlags().Lookup("common-name"))
}
