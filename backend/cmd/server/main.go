package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/container"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Print("=== umi.mikan backend started ===")

	// Create DI container
	diContainer := container.NewContainer()

	// Initialize and run server using DI container
	if err := diContainer.Invoke(runServer); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func runServer(app *container.ServerApp, cleanup *container.Cleanup) error {
	// Load port configuration
	port, err := constants.LoadPort()
	if err != nil {
		return fmt.Errorf("failed to load port: %w", err)
	}

	// Create grpc server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.AuthInterceptor),
	)

	// Register services
	g.RegisterDiaryServiceServer(grpcServer, app.DiaryService)
	g.RegisterAuthServiceServer(grpcServer, app.AuthService)
	g.RegisterUserServiceServer(grpcServer, app.UserService)

	// Enable reflection for local debugging
	// TODO: disable in production based on environment variable
	reflection.Register(grpcServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	log.Printf("gRPC server listening on :%d", port)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start gRPC server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			serverErrChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)

		// Create context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Gracefully stop the gRPC server
		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()

		// Wait for graceful stop or timeout
		select {
		case <-stopped:
			log.Print("gRPC server gracefully stopped")
		case <-ctx.Done():
			log.Print("Graceful shutdown timeout, forcing stop")
			grpcServer.Stop()
		}

		// Cleanup resources
		if err := cleanup.Close(); err != nil {
			log.Printf("Error during cleanup: %v", err)
		}

	case err := <-serverErrChan:
		log.Printf("gRPC server error: %v", err)
		// Cleanup resources on error
		if cleanupErr := cleanup.Close(); cleanupErr != nil {
			log.Printf("Error during cleanup: %v", cleanupErr)
		}
		return err
	}

	log.Print("backend end")
	return nil
}
