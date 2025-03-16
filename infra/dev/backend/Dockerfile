FROM golang:1.24
WORKDIR /backend
RUN go install github.com/air-verse/air@latest
RUN go install github.com/xo/xo@latest
CMD ["air"]
