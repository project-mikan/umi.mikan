# Goはairで常時起動なので不要
pnpm-dev:
	docker compose exec frontend pnpm dev --host 0.0.0.0
f-sh:
	docker compose exec frontend bash
f-format:
	docker compose exec frontend pnpm format

xo:
	# db-initも実行したいが立ち上げてすぐは起動できないので別でコマンド
	rm -rf backend/infrastructure/database/*.xo.go
	docker compose exec backend go tool xo schema "postgres://postgres:dev-pass@postgres/umi_mikan?sslmode=disable" -o infrastructure/database
go-mod-tidy:
	docker compose exec backend go mod tidy
# airを使うので不要↓
b-sh:
	docker compose exec backend sh
tidy:
	docker compose exec backend go mod tidy
db:
	docker compose exec postgres psql -U postgres -d umi_mikan  
db-init:
	docker compose down postgres -v
	docker compose up postgres -d
	# dbのログはすぐに取れないので別コマンドで取得する

log-f:
	docker compose logs -f frontend
log-b:
	docker compose logs -f backend
log-p:
	docker compose logs -f postgres

grpc-go:
	# 削除分は反映されないのでrm -rfしてから実行
	rm -rf backend/infrastructure/grpc/*
	protoc --go_out=backend/ \
	--go_opt=module=github.com/project-mikan/umi.mikan/backend \
	--go-grpc_out=backend/ \
	--go-grpc_opt=module=github.com/project-mikan/umi.mikan/backend \
	proto/**/*.proto

grpc-ts:
	# 削除分は反映されないのでrm -rfしてから実行
	rm -rf frontend/src/lib/grpc/*
	docker compose exec frontend pnpm dlx buf generate


grpc:
	make grpc-go
	make grpc-ts

# Backend Testing Commands
test:
	docker compose exec backend go test ./...

test-verbose:
	docker compose exec backend go test -v ./...

test-auth:
	docker compose exec backend go test -v ./service/auth

test-diary:
	docker compose exec backend go test -v ./service/diary

test-integration:
	docker compose exec backend go test -v ./test_integration

test-testkit:
	docker compose exec backend go test -v ./testkit

test-coverage:
	docker compose exec backend go test -coverprofile=coverage.out ./...
	docker compose exec backend go tool cover -html=coverage.out -o coverage.html

test-benchmark:
	docker compose exec backend go test -bench=. ./...

test-race:
	docker compose exec backend go test -race ./...

# Backend Linting Commands
lint:
	docker compose exec backend go tool golangci-lint run

lint-fix:
	docker compose exec backend go tool golangci-lint run --fix
