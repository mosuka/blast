package cmd

import (
	"context"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/mosuka/blast/client"
	"github.com/mosuka/blast/marshaler"
	"github.com/mosuka/blast/protobuf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	searchCmd = &cobra.Command{
		Use:   "search REQUEST",
		Args:  cobra.ExactArgs(1),
		Short: "Get a document",
		Long:  "Get a document",
		RunE: func(cmd *cobra.Command, args []string) error {
			grpcAddress = viper.GetString("grpc_address")

			certificateFile = viper.GetString("certificate_file")
			commonName = viper.GetString("common_name")

			searchRequest := args[0]

			m := marshaler.BlastMarshaler{}

			req := &protobuf.SearchRequest{}
			if err := m.Unmarshal([]byte(searchRequest), req); err != nil {
				return err
			}

			c, err := client.NewGRPCClientWithContextTLS(grpcAddress, context.Background(), certificateFile, commonName)
			if err != nil {
				return err
			}
			defer func() {
				_ = c.Close()
			}()

			resp, err := c.Search(req)
			if err != nil {
				return err
			}

			respBytes, err := m.Marshal(resp)
			if err != nil {
				return err
			}

			fmt.Println(string(respBytes))

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(searchCmd)

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

	searchCmd.PersistentFlags().StringVar(&configFile, "config-file", "", "config file. if omitted, blast.yaml in /etc and home directory will be searched")
	searchCmd.PersistentFlags().StringVar(&grpcAddress, "grpc-address", ":9000", "gRPC server listen address")
	searchCmd.PersistentFlags().StringVar(&certificateFile, "certificate-file", "", "path to the client server TLS certificate file")
	searchCmd.PersistentFlags().StringVar(&commonName, "common-name", "", "certificate common name")

	_ = viper.BindPFlag("grpc_address", searchCmd.PersistentFlags().Lookup("grpc-address"))
	_ = viper.BindPFlag("certificate_file", searchCmd.PersistentFlags().Lookup("certificate-file"))
	_ = viper.BindPFlag("common_name", searchCmd.PersistentFlags().Lookup("common-name"))
}
