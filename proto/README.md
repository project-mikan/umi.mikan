# proto

このファイルはdocker composeでマウントするために必要です。

## コード生成

### 事前準備（初回のみ）

```bash
brew install swift-protobuf protoc-gen-grpc-swift
```

### 生成コマンド

| コマンド | 生成先 | 説明 |
|---|---|---|
| `make grpc-go` | `backend/infrastructure/grpc/` | Go用gRPCコード |
| `make grpc-ts` | `frontend/src/lib/grpc/` | TypeScript用gRPCコード |
| `make grpc-swift` | `ios/Sources/Generated/` | Swift用gRPCコード |
| `make grpc` | 上記Go・TS両方 | Go・TSをまとめて生成 |

> `grpc-swift` はホストマシンの `protoc-gen-swift` / `protoc-gen-grpc-swift` を使用するため、brew でのインストールが必要です。
