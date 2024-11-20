package main

import (
	"context"
	"log"
	"net"

	g "github.com/project-mikan/umi.mikan/backend/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Print("=== umi.mikan backend started ===")

	// grpc のリッスンを開始
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	// サービス登録
	g.RegisterDiaryServiceServer(grpcServer, &Sample{})

	// localでcliからデバッグできるようにする
	// TODO: 環境変数で本番では有効にならないようにする
	reflection.Register(grpcServer)

	// サーバーを起動
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

	log.Print("backend end")
}

type Sample struct {
	g.UnimplementedDiaryServiceServer
}

func (s *Sample) CreateDiaryEntry(
	ctx context.Context,
	message *g.CreateDiaryEntryRequest,
) (*g.CreateDiaryEntryResponse, error) {
	return &g.CreateDiaryEntryResponse{}, nil
}

func (s *Sample) GetDiaryEntry(
	ctx context.Context,
	message *g.GetDiaryEntryRequest,
) (*g.GetDiaryEntryResponse, error) {
	return &g.GetDiaryEntryResponse{}, nil
}

func (s *Sample) ListDiaryEntries(
	ctx context.Context,
	message *g.ListDiaryEntriesRequest,
) (*g.ListDiaryEntriesResponse, error) {
	return &g.ListDiaryEntriesResponse{}, nil
}

func (s *Sample) UpdateDiaryEntry(
	ctx context.Context,
	message *g.UpdateDiaryEntryRequest,
) (*g.UpdateDiaryEntryResponse, error) {
	return &g.UpdateDiaryEntryResponse{}, nil
}

func (s *Sample) DeleteDiaryEntry(
	ctx context.Context,
	message *g.DeleteDiaryEntryRequest,
) (*g.DeleteDiaryEntryResponse, error) {
	return &g.DeleteDiaryEntryResponse{}, nil
}
