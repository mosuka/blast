package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/protobuf/ptypes/empty"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/mosuka/blast/client"
	"github.com/mosuka/blast/marshaler"
	"github.com/mosuka/blast/protobuf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	watchCmd = &cobra.Command{
		Use:   "watch",
		Short: "Watch a node updates",
		Long:  "Watch a node updates",
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

			req := &empty.Empty{}
			watchClient, err := c.Watch(req)
			if err != nil {
				return err
			}

			go func() {
				for {
					resp, err := watchClient.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						break
					}

					switch resp.Event.Type {
					case protobuf.Event_Join:
						eventReq := &protobuf.SetMetadataRequest{}
						if eventData, err := marshaler.MarshalAny(resp.Event.Data); err != nil {
							_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("%s, %v", resp.Event.Type.String(), err))
						} else {
							if eventData == nil {
								_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("%s, nil", resp.Event.Type.String()))
							} else {
								eventReq = eventData.(*protobuf.SetMetadataRequest)
							}
						}
						fmt.Printf("%s, %v\n", resp.Event.Type.String(), eventReq)
					case protobuf.Event_Leave:
						eventReq := &protobuf.DeleteMetadataRequest{}
						if eventData, err := marshaler.MarshalAny(resp.Event.Data); err != nil {
							_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("%s, %v", resp.Event.Type.String(), err))
						} else {
							if eventData == nil {
								_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("%s, nil", resp.Event.Type.String()))
							} else {
								eventReq = eventData.(*protobuf.DeleteMetadataRequest)
							}
						}
						fmt.Printf("%s, %v\n", resp.Event.Type.String(), eventReq)
					case protobuf.Event_Set:
						putRequest := &protobuf.SetRequest{}
						if putRequestInstance, err := marshaler.MarshalAny(resp.Event.Data); err != nil {
							_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("%s, %v", resp.Event.Type.String(), err))
						} else {
							if putRequestInstance == nil {
								_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("%s, nil", resp.Event.Type.String()))
							} else {
								putRequest = putRequestInstance.(*protobuf.SetRequest)
							}
						}
						fmt.Printf("%s, %v\n", resp.Event.Type.String(), putRequest)
					case protobuf.Event_Delete:
						deleteRequest := &protobuf.DeleteRequest{}
						if deleteRequestInstance, err := marshaler.MarshalAny(resp.Event.Data); err != nil {
							_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("%s, %v", resp.Event.Type.String(), err))
						} else {
							if deleteRequestInstance == nil {
								_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("%s, nil", resp.Event.Type.String()))
							} else {
								deleteRequest = deleteRequestInstance.(*protobuf.DeleteRequest)
							}
						}
						fmt.Printf("%s, %v\n", resp.Event.Type.String(), deleteRequest)
					}
				}
			}()

			quitCh := make(chan os.Signal, 1)
			signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

			<-quitCh

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(watchCmd)

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

	watchCmd.PersistentFlags().StringVar(&configFile, "config-file", "", "config file. if omitted, blast.yaml in /etc and home directory will be searched")
	watchCmd.PersistentFlags().StringVar(&grpcAddress, "grpc-address", ":9000", "gRPC server listen address")
	watchCmd.PersistentFlags().StringVar(&certificateFile, "certificate-file", "", "path to the client server TLS certificate file")
	watchCmd.PersistentFlags().StringVar(&commonName, "common-name", "", "certificate common name")

	_ = viper.BindPFlag("grpc_address", watchCmd.PersistentFlags().Lookup("grpc-address"))
	_ = viper.BindPFlag("certificate_file", watchCmd.PersistentFlags().Lookup("certificate-file"))
	_ = viper.BindPFlag("common_name", watchCmd.PersistentFlags().Lookup("common-name"))
}
