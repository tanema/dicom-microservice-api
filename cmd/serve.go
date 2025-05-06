package cmd

import (
	"log"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tanema/dicom-microservice-api/app/handlers"
	"github.com/tanema/dicom-microservice-api/app/storage"
	"github.com/tanema/dicom-microservice-api/config"
)

var (
	serveConfig *config.Config
	store       storage.Storage
)

var serveCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Run the DICON API",
	Aliases: []string{"s"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		envName, err := cmd.Flags().GetString("env")
		if err != nil {
			return err
		}
		serveConfig, err = config.Load(envName)
		if err != nil {
			return err
		}
		store, err = storage.New(serveConfig.Store)
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		http.Handle("/", handlers.NewDICOMAPI(serveConfig, store))
		log.Printf("Starting server on port %v\n", serveConfig.Port)
		return http.ListenAndServe(":"+strconv.Itoa(serveConfig.Port), nil)
	},
}
