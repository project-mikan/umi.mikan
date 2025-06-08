# umi.mikan

## getting stareted

### install

```bash
sudo pacman -S protobuf
```

メモ：go toolにしたいがdockerの外なので悩ましい

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

```bash
npm install -g grpc_tools_node_protoc_ts
```

### run

```bash
dc up -d
```

- Backend：http://localhost:8080
- Frontend：http://localhost:5173

### debug

サービス一覧

```bash
grpc_cli ls localhost:8080
```

詳細

```basrh
grpc_cli ls localhost:8080 diary.DiaryService -l
```

type表示

```bash
grpc_cli type localhost:8080 diary.CreateDiaryEntryRequest
```

remote call

```bash
grpc_cli call localhost:8080 DiaryService.CreateDiaryEntry 'title: "test",content:"test"'
```

日記検索

```bash
grpc_cli call localhost:8080 DiaryService.SearchDiaryEntries 'userID:"id" keyword:"%日記%"'
```
