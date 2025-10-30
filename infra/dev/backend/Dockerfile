FROM golang:1.25.3
WORKDIR /backend

# protoc と protoc-gen-go プラグインのインストール
RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN go install tool
CMD ["go", "tool", "air"]
