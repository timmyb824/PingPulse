# Simple multi-check uptime Dockerfile
FROM golang:1.22-alpine as builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o httping main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/httping /app/
COPY example-config.yaml /app/
EXPOSE 8080
CMD ["/app/httping", "/app/example-config.yaml"]
