package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dicom-microservice-api",
	Short: "A small DICOM API service",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	serveCmd.Flags().StringP("env", "e", "dev", "environment to run the server config")
	rootCmd.AddCommand(serveCmd)
}
