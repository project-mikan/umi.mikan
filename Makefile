pnpm-dev:
	docker compose exec frontend pnpm dev --host 0.0.0.0
f-sh:
	docker compose exec frontend bash

xo:
	docker compose exec backend xo schema "postgres://postgres:dev-pass@postgres/umi_mikan?sslmode=disable" -o infrastructure/database
go-mod-tidy:
	docker compose exec backend go mod tidy
# airを使うので不要↓
# b-dev:
# 	docker compose exec backend go run cmd/main.go
db:
	docker compose exec postgres psql -U postgres -d umi_mikan  
db-init:
	docker compose down postgres -v
	docker compose up postgres -d
	# dbのログはすぐに取れないので別コマンドで取得する

log-f:
	docker compose logs frontend
log-b:
	docker compose logs backend
log-p:
	docker compose logs postgres

grpc-go:
	protoc --go_out=backend/ \
	--go_opt=module=github.com/project-mikan/umi.mikan/backend \
	proto/hello.proto

grpc-ts:
	protoc --ts_out=grpc_js:frontend/src/lib/proto \
	-I ./proto proto/*.proto
