package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/project-mikan/umi.mikan/backend/service/auth"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
	"github.com/rs/cors"
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

	// gRPC-Webのラップ
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	// CORS設定
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// HTTPハンドラ
	handler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedGrpc.ServeHTTP(resp, req)
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: c.Handler(http.HandlerFunc(handler)),
	}

	log.Printf("gRPC-Web server listening on :%d", port)

	// HTTPサーバーを起動（gRPC-Web対応）
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

	log.Print("backend end")
}
