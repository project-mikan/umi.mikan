package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/container"
	connectadapter "github.com/project-mikan/umi.mikan/backend/infrastructure/connectrpc"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/grpc/grpcconnect"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/mcpserver"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Print("=== umi.mikan backend started ===")

	// Create DI container
	diContainer, err := container.NewContainer()
	if err != nil {
		log.Fatalf("Failed to create DI container: %v", err)
	}

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
	g.RegisterEntityServiceServer(grpcServer, app.EntityService)
	g.RegisterUserServiceServer(grpcServer, app.UserService)

	// Enable reflection based on environment variable
	if constants.LoadGRPCReflectionEnabled() {
		log.Print("gRPC reflection enabled")
		reflection.Register(grpcServer)
	} else {
		log.Print("gRPC reflection disabled")
	}

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	log.Printf("gRPC server listening on :%d", port)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Prometheusメトリクスサーバーを起動（デバッグエンドポイント含む）
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/debug/error", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Error: debug test error triggered")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"ok":true}`)); err != nil {
			log.Printf("Error: failed to write debug response: %v", err)
		}
	})
	mux.HandleFunc("/debug/warn", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Warn: debug test warning triggered")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"ok":true}`)); err != nil {
			log.Printf("Error: failed to write debug response: %v", err)
		}
	})
	metricsServer := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}
	go func() {
		log.Print("Metrics server listening on :8082")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// ConnectRPC HTTP サーバーを起動（iOS/外部クライアント向け）
	connectMux := http.NewServeMux()
	authInterceptor := connect.WithInterceptors(connectadapter.NewAuthInterceptor())
	connectMux.Handle(grpcconnect.NewAuthServiceHandler(connectadapter.NewAuthServiceAdapter(app.AuthService), authInterceptor))
	connectMux.Handle(grpcconnect.NewDiaryServiceHandler(connectadapter.NewDiaryServiceAdapter(app.DiaryService), authInterceptor))
	connectMux.Handle(grpcconnect.NewEntityServiceHandler(connectadapter.NewEntityServiceAdapter(app.EntityService), authInterceptor))
	connectMux.Handle(grpcconnect.NewUserServiceHandler(connectadapter.NewUserServiceAdapter(app.UserService), authInterceptor))
	// Protocols フィールドで HTTP/1.1 と HTTP/2 をクリアテキスト（h2c）で有効にする。
	// 本番環境では Cloudflare がTLS終端するため、バックエンドはプレーン HTTP で受け取る。
	connectServer := &http.Server{
		Addr:      ":8013",
		Handler:   connectMux,
		Protocols: new(http.Protocols),
	}
	connectServer.Protocols.SetHTTP1(true)
	connectServer.Protocols.SetUnencryptedHTTP2(true)

	// MCP（Model Context Protocol）サーバーを起動
	// AIクライアント（Claude Desktopなど）向けに日記取得・検索ツールを公開する
	mcpServer := &http.Server{
		Addr:    ":8014",
		Handler: mcpserver.NewHTTPHandler(app.DiaryService),
	}

	// Start gRPC server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			serverErrChan <- err
		}
	}()
	// ConnectRPC サーバーの起動エラーも同じチャンネルで検知する
	go func() {
		log.Print("ConnectRPC server listening on :8013")
		if err := connectServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("ConnectRPC server error: %v", err)
			serverErrChan <- err
		}
	}()
	// MCP サーバーの起動エラーも同じチャンネルで検知する
	go func() {
		log.Print("MCP server listening on :8014")
		if err := mcpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("MCP server error: %v", err)
			serverErrChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)

		// gRPC GracefulStop 用のコンテキスト（最大 30 秒）
		grpcCtx, grpcCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer grpcCancel()

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
		case <-grpcCtx.Done():
			log.Print("Graceful shutdown timeout, forcing stop")
			grpcServer.Stop()
		}

		// HTTP サーバー停止用のコンテキストを独立して生成する。
		// gRPC の停止で時間を消費しても ConnectRPC/メトリクスが十分な猶予を持てるようにする。
		httpCtx, httpCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer httpCancel()

		// ConnectRPC サーバーを停止
		if err := connectServer.Shutdown(httpCtx); err != nil {
			log.Printf("Error shutting down ConnectRPC server: %v", err)
		}

		// MCP サーバーを停止
		if err := mcpServer.Shutdown(httpCtx); err != nil {
			log.Printf("Error shutting down MCP server: %v", err)
		}

		// メトリクスサーバーを停止
		if err := metricsServer.Shutdown(httpCtx); err != nil {
			log.Printf("Error shutting down metrics server: %v", err)
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
