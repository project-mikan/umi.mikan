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
	docker compose exec backend go fix ./...
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
	docker compose exec backend go tool dbtpl schema "postgres://postgres:dev-pass@postgres/umi_mikan?sslmode=disable" -o infrastructure/database
	# pgvectorの拡張関数(sf_*)はコード生成対象外のため削除
	rm -f backend/infrastructure/database/sf_*.dbtpl.go
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
	# Apply schema changes to database (本番DB)
	docker compose exec backend go tool pg-schema-diff apply \
		--from-dsn "postgres://postgres:dev-pass@postgres/umi_mikan?sslmode=disable" \
		--to-dir /schema \
		--disable-plan-validation \
		--allow-hazards DELETES_DATA,INDEX_DROPPED,INDEX_BUILD,ACQUIRES_ACCESS_EXCLUSIVE_LOCK \
		--skip-confirm-prompt

db-apply-test:
	# Apply schema changes to test database (テストDB)
	docker compose exec backend go tool pg-schema-diff apply \
		--from-dsn "postgres://postgres:test-pass@postgres_test/umi_mikan_test?sslmode=disable" \
		--to-dir /schema \
		--disable-plan-validation \
		--allow-hazards DELETES_DATA,INDEX_DROPPED,INDEX_BUILD,ACQUIRES_ACCESS_EXCLUSIVE_LOCK,UPGRADING_EXTENSION_VERSION \
		--skip-confirm-prompt

f-log:
	docker compose logs -f frontend
b-log:
	docker compose logs -f backend
p-log:
	docker compose logs -f postgres

grpc-go:
	# 削除分は反映されないのでrm -rfしてから実行
	rm -rf backend/infrastructure/grpc/*
	# protoc-gen-connect-go は protoc-gen-connect-go@latest で /go/bin に入れておく
	docker compose exec backend protoc --proto_path=/proto \
	--go_out=. \
	--go_opt=module=github.com/project-mikan/umi.mikan/backend \
	--go-grpc_out=. \
	--go-grpc_opt=module=github.com/project-mikan/umi.mikan/backend \
	--connect-go_out=. \
	--connect-go_opt=module=github.com/project-mikan/umi.mikan/backend \
	--plugin=protoc-gen-connect-go=/go/bin/protoc-gen-connect-go \
	auth/auth.proto diary/diary.proto user/user.proto entity/entity.proto

grpc-ts:
	# 削除分は反映されないのでrm -rfしてから実行
	rm -rf frontend/src/lib/grpc/*
	docker compose exec frontend pnpm exec buf generate
	docker compose exec frontend pnpm format


# まとめてやると↓なんか動かない？
grpc-swift:
	# 削除分は反映されないのでrm -rfしてから実行
	rm -rf ios/Sources/Proto/*
	mkdir -p ios/Sources/Proto
	# connect-swift用: protoc-gen-swift（protobufメッセージ）+ protoc-gen-connect-swift（ConnectRPCクライアント）
	# protoc-gen-connect-swift は connect-swift のリリースページからダウンロードして brew --prefix/bin に配置する
	# https://github.com/connectrpc/connect-swift/releases
	protoc --proto_path=proto \
	--swift_out=ios/Sources/Proto \
	--connect-swift_out=ios/Sources/Proto \
	--plugin=protoc-gen-connect-swift=$(shell brew --prefix)/bin/protoc-gen-connect-swift \
	proto/auth/auth.proto proto/diary/diary.proto proto/user/user.proto proto/entity/entity.proto

grpc:
	make grpc-go
	make grpc-ts
	make grpc-swift

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

# GEMINI_API_KEY_FOR_TEST=xxx make b-test-semantic-eval で実行する
b-test-semantic-eval:
	docker compose exec -e GEMINI_API_KEY_FOR_TEST=$(GEMINI_API_KEY_FOR_TEST) backend go test -tags=integration -v -run TestSemanticSearchEvaluation ./infrastructure/database/...

b-test-race:
	docker compose exec backend go test -race ./...

# iOS
IOS_PROJECT = ios/umi.mikan.xcodeproj
IOS_SCHEME = umi.mikan
IOS_DESTINATION = platform=iOS Simulator,name=iPhone 17

ios-format:
	cd ios && swiftformat --config .swiftformat Sources/ Shared/ Widgets/

ios-lint:
	make ios-format
	cd ios && swiftlint lint --config .swiftlint.yml

ios-build:
	xcodebuild build -project $(IOS_PROJECT) -scheme "$(IOS_SCHEME)" -destination "$(IOS_DESTINATION)" -quiet

ios-test:
	xcodebuild test -project $(IOS_PROJECT) -scheme "$(IOS_SCHEME)" -destination "$(IOS_DESTINATION)" -quiet

ios-log:
	xcrun simctl spawn booted log stream --predicate 'processImagePath contains "umi.mikan"' 2>/dev/null || echo "アプリが起動していません"

1:
	make b-lint
	make f-lint
	make ios-lint
	make b-test
	make f-test
