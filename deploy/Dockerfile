# ============ 构建阶段 ============
FROM golang:1.25-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# ============ 运行阶段 ============
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server /app/server
COPY migrations /app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/server"]
