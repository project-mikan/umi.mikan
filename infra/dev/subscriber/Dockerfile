FROM golang:1.25.1
WORKDIR /backend
RUN go install tool
CMD ["go", "run", "cmd/subscriber/main.go"]