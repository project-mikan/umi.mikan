# umi.mikan

## getting stareted

### install

```bash
sudo pacman -S protobuf
```

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### gen

```bash
protoc --go_out=backend/ --go-grpc_out=backend/ proto/hello.proto
```
