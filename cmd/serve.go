package cmd

import (
	"go.uber.org/zap"

	"github.com/liuyoshio/platformd/internal/server"
	"github.com/spf13/cobra"
)

var grpcPort int
var httpPort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the platformd API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		log, _ := zap.NewProduction()
		if verbose {
			log, _ = zap.NewDevelopment()
		}
		defer log.Sync()
		return server.Run(grpcPort, httpPort, log)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVar(&grpcPort, "grpc-port", 50051, "gRPC listen port")
	serveCmd.Flags().IntVar(&httpPort, "http-port", 8080, "REST listen port")
}
