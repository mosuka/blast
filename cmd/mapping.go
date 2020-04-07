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
	mappingCmd = &cobra.Command{
		Use:   "mapping",
		Short: "Get the index mapping",
		Long:  "Get the index mapping",
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

			resp, err := c.Mapping()
			if err != nil {
				return err
			}

			fmt.Println(string(resp.Mapping))

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(mappingCmd)

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

	mappingCmd.PersistentFlags().StringVar(&configFile, "config-file", "", "config file. if omitted, blast.yaml in /etc and home directory will be searched")
	mappingCmd.PersistentFlags().StringVar(&grpcAddress, "grpc-address", ":9000", "gRPC server listen address")
	mappingCmd.PersistentFlags().StringVar(&certificateFile, "certificate-file", "", "path to the client server TLS certificate file")
	mappingCmd.PersistentFlags().StringVar(&commonName, "common-name", "", "certificate common name")

	_ = viper.BindPFlag("grpc_address", mappingCmd.PersistentFlags().Lookup("grpc-address"))
	_ = viper.BindPFlag("certificate_file", mappingCmd.PersistentFlags().Lookup("certificate-file"))
	_ = viper.BindPFlag("common_name", mappingCmd.PersistentFlags().Lookup("common-name"))
}
