# 部署到公共网络环境

本文档说明如何把交通数据分析平台部署到一台有公网 IP 的云服务器。

## 架构

```
                 ┌─────────────────────────────────────────┐
                 │              云服务器                    │
  公网 ────────▶ │  frontend (Nginx :80/:443)               │
  浏览器/客户端   │    ├─ 静态文件（React 构建产物）          │
                 │    └─ /api/* 反向代理 ──▶ backend (:8080) │
                 │                              │           │
                 │                              ▼           │
                 │                     postgres (:5432,内网) │
                 └─────────────────────────────────────────┘
```

- 只有 `frontend` 容器暴露公网端口，`backend` 和 `postgres` 只在 Docker 内网，安全性更好。
- 前端与后端同域，前端代码里的 `/api` 由 Nginx 反向代理到后端，无需改前端代码。

## 前置条件

1. **云服务器**：有公网 IP，推荐 2 核 2G 以上（阿里云 / 腾讯云轻量应用服务器 / AWS / DigitalOcean 等均可）。
2. **Docker + Docker Compose**：已安装（服务器上执行 `docker --version` 确认）。
3. **（可选）域名**：HTTPS 需要域名，纯 IP 只能用 HTTP。

## 部署步骤

### 1. 上传代码到服务器

```bash
# 方式一：git 仓库（推荐）
git clone <你的仓库地址> traffic-platform
cd traffic-platform

# 方式二：本地打包上传
# 本地：tar -czf traffic-platform.tar.gz --exclude=frontend/node_modules --exclude=frontend/dist .
# 服务器：scp traffic-platform.tar.gz user@服务器IP:/root && tar -xzf traffic-platform.tar.gz
```

### 2. 设置数据库密码（强烈建议）

```bash
# 生产环境务必使用强密码，不要用默认的 traffic
export POSTGRES_PASSWORD="你的强密码"
```

### 3. 构建并启动

```bash
docker compose -f docker-compose.prod.yml up -d --build
```

首次构建会下载依赖并编译（Go 后端 + 前端），约需几分钟。

### 4. 验证

```bash
# 健康检查
curl http://localhost/healthz
# 期望返回 {"code":0,"message":"ok","data":{"status":"ok"}}

# 访问前端
# 浏览器打开 http://服务器IP
```

### 5. 云服务商安全组放行端口

在云控制台的「安全组 / 防火墙」放行：
- `80`（HTTP）
- `443`（HTTPS，如启用）

## 常用运维命令

```bash
# 查看服务状态
docker compose -f docker-compose.prod.yml ps

# 查看后端日志
docker compose -f docker-compose.prod.yml logs -f backend

# 重启
docker compose -f docker-compose.prod.yml restart

# 停止
docker compose -f docker-compose.prod.yml down
```

## 更新部署（代码有改动后）

```bash
cd traffic-platform
git pull                      # 拉取最新代码
docker compose -f docker-compose.prod.yml up -d --build
```

## 配置 HTTPS（可选，需要域名）

用 certbot 自动签发 Let's Encrypt 免费证书：

```bash
# 1. 安装 certbot
apt install -y certbot python3-certbot-nginx   # Ubuntu/Debian

# 2. 先在 nginx.conf 把 server_name 改成你的域名，再签发
certbot --nginx -d 你的域名
```

certbot 会自动修改 Nginx 配置并续期证书。也可以在云服务商控制台申请免费 SSL 证书后手动挂载到 `frontend` 容器的 443 端口。

## 常见问题

| 问题 | 处理 |
|------|------|
| 后端起不来（连不上数据库） | `docker compose logs backend`，确认 `POSTGRES_PASSWORD` 与 postgres 容器一致 |
| 首次建表未执行 | 挂载的 `001_init.sql` 只在**数据卷为空**时执行；已建过则手动跑 `002_aggregation_idempotency.sql` |
| 前端 404（刷新页面） | 已配置 SPA 回退 `try_files ... /index.html`，正常 |
| 想只暴露 HTTPS | 安全组只放行 443，`frontend` 只映射 `443:443` |

## 生产环境建议

- `POSTGRES_PASSWORD` 用强随机密码，并通过环境变量或 `.env` 注入（不要写死在 compose 里）。
- 定期备份数据库：`docker exec traffic-platform-postgres-1 pg_dump -U traffic traffic > backup.sql`（容器名以 `docker compose ps` 为准）。
- 后端可用 `FREE_FLOW_SPEED`、`LOW_SPEED_THRESHOLD`、`FLOW_SPIKE_FACTOR`、`LOW_SPEED_CONSECUTIVE` 环境变量调整阈值。
