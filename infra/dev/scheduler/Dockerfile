FROM golang:1.25.1
WORKDIR /backend
RUN go install tool
CMD ["go", "tool", "air", "-c", ".air-scheduler.toml"]