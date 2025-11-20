# フロントエンド
f-sh:
	docker compose exec frontend bash
f-format:
	docker compose exec frontend pnpm format
f-build:
	docker build -t umi-mikan-frontend-test:0.0.1 -f ./infra/prod/frontend/Dockerfile ./frontend 
f-build-no-cache:
	docker build -t umi-mikan-frontend-test:0.0.1 --no-cache -f ./infra/prod/frontend/Dockerfile ./frontend 

f-lint:
	make f-format
	docker compose exec frontend pnpm 1

# バックエンド
b-format:
	docker compose exec backend go fmt ./...
	docker compose exec backend go tool golangci-lint run --fix
b-lint:
	make b-format
	docker compose exec backend go tool golangci-lint run
b-sh:
	docker compose exec backend sh
tidy:
	docker compose exec backend go mod tidy


xo:
	# db-initを別で実行することでDBを更新できる
	rm -rf backend/infrastructure/database/*.xo.go
	docker compose exec backend go tool xo schema "postgres://postgres:dev-pass@postgres/umi_mikan?sslmode=disable" -o infrastructure/database
go-mod-tidy:
	docker compose exec backend go mod tidy
# airを使うので不要↓
db:
	docker compose exec postgres psql -U postgres -d umi_mikan
# db-init:
# 	docker compose down postgres -v
# 	docker compose up postgres -d
	# dbのログはすぐに取れないので別コマンドで取得する

db-diff:
	# pg-schema-diff dry run - show what changes would be made
	docker compose exec backend go tool pg-schema-diff plan \
		--from-dsn "postgres://postgres:dev-pass@postgres/umi_mikan?sslmode=disable" \
		--to-dir /schema \
		--disable-plan-validation

db-apply:
	# Apply schema changes to database
	docker compose exec backend go tool pg-schema-diff apply \
		--from-dsn "postgres://postgres:dev-pass@postgres/umi_mikan?sslmode=disable" \
		--to-dir /schema \
		--disable-plan-validation \
		--allow-hazards DELETES_DATA,INDEX_DROPPED

f-log:
	docker compose logs -f frontend
b-log:
	docker compose logs -f backend
p-log:
	docker compose logs -f postgres

grpc-go:
	# 削除分は反映されないのでrm -rfしてから実行
	rm -rf backend/infrastructure/grpc/*
	docker compose exec backend protoc --proto_path=/proto \
	--go_out=. \
	--go_opt=module=github.com/project-mikan/umi.mikan/backend \
	--go-grpc_out=. \
	--go-grpc_opt=module=github.com/project-mikan/umi.mikan/backend \
	auth/auth.proto diary/diary.proto user/user.proto entity/entity.proto

grpc-ts:
	# 削除分は反映されないのでrm -rfしてから実行
	rm -rf frontend/src/lib/grpc/*
	docker compose exec frontend pnpm exec buf generate
	docker compose exec frontend pnpm format


grpc:
	make grpc-go
	make grpc-ts

f-test:
	docker compose exec frontend pnpm test:run
b-test:
	docker compose exec backend go test ./...

b-test-verbose:
	docker compose exec backend go test -v ./...

b-test-auth:
	docker compose exec backend go test -v ./service/auth

b-test-diary:
	docker compose exec backend go test -v ./service/diary

b-test-integration:
	docker compose exec backend go test -v ./test_integration

b-test-testkit:
	docker compose exec backend go test -v ./testkit

b-test-coverage:
	docker compose exec backend go test -coverprofile=coverage.out ./...
	docker compose exec backend go tool cover -html=coverage.out -o coverage.html

b-test-benchmark:
	docker compose exec backend go test -bench=. ./...

b-test-race:
	docker compose exec backend go test -race ./...
1:
	make b-lint
	make f-lint
	make b-test
	make f-test
