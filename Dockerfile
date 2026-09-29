# ============ 阶段1：前端构建 ============
FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# ============ 阶段2：后端构建 ============
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# ============ 阶段3：运行（后端 + 前端静态文件）============
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=backend /server /app/server
COPY --from=frontend /app/dist /app/frontend/dist
COPY migrations /app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/server"]
