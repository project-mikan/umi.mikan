f-dev:
	docker compose exec frontend pnpm dev --host 0.0.0.0
# airを使うので不要↓
# b-dev:
# 	docker compose exec backend go run cmd/main.go
db:
	docker compose exec postgres psql -U postgres -d umi_mikan  
