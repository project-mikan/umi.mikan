FROM golang:1.24.5-alpine AS builder

WORKDIR /build

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates git

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o binary cmd/server/main.go

# Production stage
FROM scratch

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /build/binary /binary

EXPOSE 8080

CMD ["/binary"]
