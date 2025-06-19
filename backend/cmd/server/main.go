package main

import (
	"fmt"
	"log"
	"net"

	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/project-mikan/umi.mikan/backend/service/auth"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
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
	defer db.Close()

	// サービス登録
	g.RegisterDiaryServiceServer(grpcServer, &diary.DiaryEntry{DB: db})
	g.RegisterAuthServiceServer(grpcServer, &auth.AuthEntry{DB: db})

	// localでcliからデバッグできるようにする
	// TODO: 環境変数で本番では有効にならないようにする
	reflection.Register(grpcServer)

	// gRPCサーバーを起動
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on :%d", port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

	log.Print("backend end")
}
