package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/liuyoshio/platformd/internal/catalog"
	catalogv1 "github.com/liuyoshio/platformd/proto/catalogv1"
)

func Run(grpcPort, httpPort int, log *zap.Logger) error {
	store := catalog.NewStore()
	srv := NewCatalogServer(store)

	// --- gRPC server ---
	grpcServer := grpc.NewServer()
	catalogv1.RegisterCatalogServiceServer(grpcServer, srv)
	reflection.Register(grpcServer) // ← 加这行

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	go func() {
		log.Info("gRPC listening", zap.Int("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("grpc serve", zap.Error(err))
		}
	}()

	// --- REST gateway (talks to the gRPC server over loopback) ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	err = catalogv1.RegisterCatalogServiceHandlerFromEndpoint(
		ctx, mux, fmt.Sprintf("localhost:%d", grpcPort),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		return fmt.Errorf("register gateway: %w", err)
	}
	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", httpPort), Handler: mux}
	go func() {
		log.Info("REST listening", zap.Int("port", httpPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http serve", zap.Error(err))
		}
	}()

	// --- graceful shutdown on SIGINT/SIGTERM ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("shutting down")

	shutdownCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
	defer c()
	_ = httpServer.Shutdown(shutdownCtx)
	grpcServer.GracefulStop()
	return nil
}
