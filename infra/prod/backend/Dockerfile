FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine AS builder

ARG TARGETARCH

WORKDIR /build

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates git

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build \
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
