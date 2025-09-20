# Pub/sub

非同期で実行できる仕組み

## 機能概要

- 非同期で実行開始・処理できる仕組みを作る
- 外部依存でなくdocker composeで実行できるものとする
- 用途はLLMの要約生成など時間を要するものを、日記の保存などをトリガーとして実行するため

## 技術

- メッセージキューイングはRedis Pub/Subを用いる
- Publisheはbackend/cmd/sever内部で
- Subscribeはbackend/cmd/subscriberに実装し、Goで処理を行う

## 動作
