package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Print("=== umi.mikan backend started ===")

	port, err := constants.LoadPort()
	if err != nil {
		log.Fatalf(err.Error())
	}
	// grpc のリッスンを開始
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	// DB接続
	dbConfig, err := constants.LoadDBConfig()
	if err != nil {
		log.Fatalf(err.Error())
	}
	db := database.NewDB(dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)
	defer db.Close()

	// サービス登録
	g.RegisterDiaryServiceServer(grpcServer, &DiaryEntyry{db: db})

	// localでcliからデバッグできるようにする
	// TODO: 環境変数で本番では有効にならないようにする
	reflection.Register(grpcServer)

	// サーバーを起動
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

	log.Print("backend end")
}

type DiaryEntyry struct {
	g.UnimplementedDiaryServiceServer
	db database.DB
}

func (s *DiaryEntyry) CreateDiaryEntry(
	ctx context.Context,
	message *g.CreateDiaryEntryRequest,
) (*g.CreateDiaryEntryResponse, error) {
	return &g.CreateDiaryEntryResponse{}, nil
}

func (s *DiaryEntyry) GetDiaryEntry(
	ctx context.Context,
	message *g.GetDiaryEntryRequest,
) (*g.GetDiaryEntryResponse, error) {
	return &g.GetDiaryEntryResponse{}, nil
}

func (s *DiaryEntyry) ListDiaryEntries(
	ctx context.Context,
	message *g.ListDiaryEntriesRequest,
) (*g.ListDiaryEntriesResponse, error) {
	return &g.ListDiaryEntriesResponse{}, nil
}

func (s *DiaryEntyry) UpdateDiaryEntry(
	ctx context.Context,
	message *g.UpdateDiaryEntryRequest,
) (*g.UpdateDiaryEntryResponse, error) {
	return &g.UpdateDiaryEntryResponse{}, nil
}

func (s *DiaryEntyry) DeleteDiaryEntry(
	ctx context.Context,
	message *g.DeleteDiaryEntryRequest,
) (*g.DeleteDiaryEntryResponse, error) {
	return &g.DeleteDiaryEntryResponse{}, nil
}

func (s *DiaryEntyry) SearchDiaryEntries(
	ctx context.Context,
	message *g.SearchDiaryEntriesRequest,
) (*g.SearchDiaryEntriesResponse, error) {
	ds, err := database.DiariesByUserIDAndContent(ctx, s.db, message.UserID, message.Keyword)
	if err != nil {
		return nil, err
	}

	fmt.Printf("SearchDiariesEntry: %v", message)
	fmt.Printf("Response: %v", ds)
	return &g.SearchDiaryEntriesResponse{
		Entries: nil,
	}, nil
}
