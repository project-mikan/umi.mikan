FROM golang:1.25.3
WORKDIR /backend
RUN go install github.com/air-verse/air@latest
CMD ["air", "-c", ".air-scheduler.toml"]