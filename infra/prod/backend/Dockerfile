FROM golang:1.24.4-alpine as builder

WORKDIR /build

RUN apk update && apk add git
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build cmd/server/main.go -o binary


FROM scratch as prod

WORKDIR /backend
COPY --from=builder /build/binary binary

EXPOSE 8080
CMD ["/backend/binary"]
