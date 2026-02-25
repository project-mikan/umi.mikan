FROM golang:1.26.0
WORKDIR /backend
RUN go install github.com/air-verse/air@latest
CMD ["air", "-c", ".air-scheduler.toml"]