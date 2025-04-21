FROM golang:1.22-alpine as builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o pingpulse

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/pingpulse /app/
COPY example-config.yaml /app/
RUN chmod +x /app/pingpulse
EXPOSE 8080
CMD ["/app/pingpulse", "/app/example-config.yaml"]
