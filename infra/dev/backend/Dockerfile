FROM golang:1.23
WORKDIR /backend
RUN go install github.com/air-verse/air@latest
RUN go install github.com/xo/xo@latest
CMD ["air"]
