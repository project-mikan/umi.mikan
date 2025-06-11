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
	npx grpc_tools_node_protoc \
  	--proto_path=./proto \
  	--js_out=import_style=commonjs,binary:./frontend/src/lib/grpc \
  	--grpc_out=grpc_js:./frontend/src/lib/grpc \
  	--ts_out=grpc_js:./frontend/src/lib/grpc \
  	./proto/**/*.proto


grpc:
	make grpc-go
	make grpc-ts
