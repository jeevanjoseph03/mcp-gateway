FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o mcp-gateway cmd/gateway/main_metrics.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/mcp-gateway .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./mcp-gateway", "-config", "configs/config.yaml"]
