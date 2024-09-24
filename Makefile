f-dev:
	docker compose exec frontend pnpm dev --host 0.0.0.0
f-sh:
	docker compose exec frontend bash
# airを使うので不要↓
# b-dev:
# 	docker compose exec backend go run cmd/main.go
db:
	docker compose exec postgres psql -U postgres -d umi_mikan  

grpc-go:
	protoc --go_out=backend/ \
	--go_opt=module=github.com/project-mikan/umi.mikan/backend \
	proto/hello.proto

grpc-ts:
	protoc --ts_out=grpc_js:frontend/src/lib/proto \
	-I ./proto proto/*.proto
