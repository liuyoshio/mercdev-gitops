package cmd

import (
	"os"

	"go.uber.org/zap"

	"github.com/liuyoshio/platformd/internal/server"
	"github.com/spf13/cobra"
)

var grpcPort int
var httpPort int
var otelEndpoint string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the platformd API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		log, _ := zap.NewProduction()
		if verbose {
			log, _ = zap.NewDevelopment()
		}
		defer log.Sync()

		endpoint := otelEndpoint
		if endpoint == "" {
			endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		}
		if endpoint == "" {
			endpoint = "localhost:4317"
		}
		return server.Run(grpcPort, httpPort, endpoint, log)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVar(&grpcPort, "grpc-port", 50051, "gRPC listen port")
	serveCmd.Flags().IntVar(&httpPort, "http-port", 8080, "REST listen port")
	serveCmd.Flags().StringVar(&otelEndpoint, "otel-endpoint", "", "OTLP gRPC endpoint for traces")
}
