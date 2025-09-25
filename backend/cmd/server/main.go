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
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/project-mikan/umi.mikan/backend/service/auth"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
	"github.com/project-mikan/umi.mikan/backend/service/user"
	"github.com/redis/rueidis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Print("=== umi.mikan backend started ===")

	port, err := constants.LoadPort()
	if err != nil {
		log.Fatalf("%v", err)
	}

	// grpc server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.AuthInterceptor),
	)

	// DB接続
	dbConfig, err := constants.LoadDBConfig()
	if err != nil {
		log.Fatalf("%v", err)
	}
	db := database.NewDB(dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	// Redis接続
	redisConfig, err := constants.LoadRedisConfig()
	if err != nil {
		log.Fatalf("Failed to load Redis config: %v", err)
	}

	redisClient, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)},
	})
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer redisClient.Close()

	// サービス登録
	g.RegisterDiaryServiceServer(grpcServer, &diary.DiaryEntry{DB: db, Redis: redisClient})
	g.RegisterAuthServiceServer(grpcServer, &auth.AuthEntry{DB: db})
	g.RegisterUserServiceServer(grpcServer, &user.UserEntry{DB: db, RedisClient: redisClient})

	// localでcliからデバッグできるようにする
	// TODO: 環境変数で本番では有効にならないようにする
	reflection.Register(grpcServer)

	// gRPCサーバーを起動
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
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

	case err := <-serverErrChan:
		log.Printf("gRPC server error: %v", err)
	}

	log.Print("backend end")
}
