package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "blast",
		Short: "The lightweight distributed search server",
		Long:  "The lightweight distributed search server",
	}
)

func Execute() error {
	return rootCmd.Execute()
}
