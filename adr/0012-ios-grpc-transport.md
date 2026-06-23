# ADR 0012: iOSアプリのバックエンド通信プロトコル選定

## ステータス
Accepted

## コンテキスト

### モチベーション

- iOSアプリからバックエンドAPIに接続するため、既存のgRPCエンドポイントを外部公開する必要がある
- バックエンドはすでにCloudflare Tunnel（Public hostname方式）で運用中
- 追加のインフラ変更を最小限に抑えたい

### 調査した選択肢

#### 案1: gRPC over Cloudflare Tunnel Public hostname

ログ調査でTLSハンドシェイク（TLS 1.3、ALPN h2）は成功することを確認したが、
Cloudflareが`application/grpc` Content-Typeのトラフィックをブロックすることが判明した。

```
# iOSログより: TLS接続成功後すぐにキャンセルされる
alpn(h2) ... Duration: 0.112s ... cancelled
bytes in/out: 5698/2017
```

Cloudflare公式ドキュメントにも「Public hostname deployments are not currently supported」と明記されている。

→ **採用不可**

#### 案2: VPSのTCPポートを直接公開 + TLS

- サーバーのファイアウォールでgRPC用ポートを開放
- Let's Encryptで証明書を取得してバックエンドに設定
- Cloudflare不使用

→ インフラ管理の複雑さが増すため見送り

#### 案3: gRPC-Web

- HTTP/1.1で動くgRPCのサブセット
- iOS Swift向けの公式クライアントライブラリが存在しない

→ **採用不可**

#### 案4: ConnectRPC（採用）

- Bufチームが開発するgRPC互換プロトコル
- 同一の`.proto`定義をそのまま使用可能
- HTTP/1.1およびHTTP/2の両方に対応
- `Content-Type: application/proto`を使用するためCloudflareにブロックされない
- Go向け: `connectrpc.com/connect`
- iOS向け: `connect-swift`（公式ライブラリあり）

**ストリーミング対応:**

| 種別 | HTTP/1.1 | HTTP/2（Cloudflare経由） |
|---|---|---|
| Unary | ✅ | ✅ |
| サーバーストリーミング | ✅ | ✅ |
| クライアントストリーミング | ✅ | ✅ |
| 双方向ストリーミング | ❌ | ✅ |

CloudflareはHTTP/2自体は通すため、ConnectRPCを使えば双方向ストリーミングを含む全機能が利用可能。
umi.mikanは現状リアルタイム双方向通信を必要としないため、HTTP/1.1フォールバックの制限も問題なし。

## 決定

iOSアプリのバックエンド通信プロトコルとして**ConnectRPC**を採用する。

### アーキテクチャ概要

```
iOS App (connect-swift)
    ↓ HTTPS (TLS終端: Cloudflare)
umi-mikan-api.usuyuki.net
    ↓ HTTP (Cloudflare Tunnel)
backend:8080 (connectrpc.com/connect)
    ↓
既存gRPCサービス実装（変更なし）
```

### 実装方針

- バックエンド: 既存の`DiaryService`・`AuthService`のgRPCハンドラをConnectRPCでラップ
- `.proto`ファイルの変更は不要
- iOS: `connect-swift`でクライアントコードを生成し、`GRPCClient.swift`を置き換え
- 既存のWebフロントエンド（SvelteKit）はgRPCのままで変更なし

## 結果

- 追加サーバー公開ポート不要、既存Cloudflare Tunnelをそのまま活用できる
- `.proto`定義の二重管理が不要
- 将来的にリアルタイム機能が必要になった場合もConnectRPCのHTTP/2対応で対処可能
