package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/mosuka/blast/client"
	"github.com/mosuka/blast/protobuf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	bulkDeleteCmd = &cobra.Command{
		Use:   "bulk-delete",
		Short: "Delete a document",
		Long:  "Delete a document",
		RunE: func(cmd *cobra.Command, args []string) error {
			grpcAddress = viper.GetString("grpc_address")

			certificateFile = viper.GetString("certificate_file")
			commonName = viper.GetString("common_name")

			req := &protobuf.BulkDeleteRequest{
				Requests: make([]*protobuf.DeleteRequest, 0),
			}

			var reader *bufio.Reader
			if file != "" {
				// from file
				f, err := os.Open(file)
				if err != nil {
					return err
				}
				defer f.Close()
				reader = bufio.NewReader(f)
			} else {
				// from stdin
				reader = bufio.NewReader(os.Stdin)
			}

			for {
				docBytes, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF || err == io.ErrClosedPipe {
						if len(docBytes) > 0 {
							r := &protobuf.DeleteRequest{
								Id: strings.TrimSpace(string(docBytes)),
							}
							req.Requests = append(req.Requests, r)
						}
						break
					}
				}
				if len(docBytes) > 0 {
					r := &protobuf.DeleteRequest{
						Id: strings.TrimSpace(string(docBytes)),
					}
					req.Requests = append(req.Requests, r)
				}
			}

			c, err := client.NewGRPCClientWithContextTLS(grpcAddress, context.Background(), certificateFile, commonName)
			if err != nil {
				return err
			}
			defer func() {
				_ = c.Close()
			}()

			count, err := c.BulkDelete(req)
			if err != nil {
				return err
			}

			fmt.Println(count)

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(bulkDeleteCmd)

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

	bulkDeleteCmd.PersistentFlags().StringVar(&configFile, "config-file", "", "config file. if omitted, blast.yaml in /etc and home directory will be searched")
	bulkDeleteCmd.PersistentFlags().StringVar(&grpcAddress, "grpc-address", ":9000", "gRPC server listen address")
	bulkDeleteCmd.PersistentFlags().StringVar(&certificateFile, "certificate-file", "", "path to the client server TLS certificate file")
	bulkDeleteCmd.PersistentFlags().StringVar(&commonName, "common-name", "", "certificate common name")

	bulkDeleteCmd.PersistentFlags().StringVar(&file, "file", "", "path to the file that documents have written in NDJSON(JSONL) format")

	_ = viper.BindPFlag("grpc_address", bulkDeleteCmd.PersistentFlags().Lookup("grpc-address"))
	_ = viper.BindPFlag("certificate_file", bulkDeleteCmd.PersistentFlags().Lookup("certificate-file"))
	_ = viper.BindPFlag("common_name", bulkDeleteCmd.PersistentFlags().Lookup("common-name"))
}
