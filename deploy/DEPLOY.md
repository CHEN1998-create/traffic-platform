# 部署到公共网络环境

本文提供两种部署方案：

- **方案 A：Render（PaaS，无需服务器，推荐）** —— 从 Git 仓库一键部署，免费层可用，最适合没有云服务器的场景。
- **方案 B：自有云服务器 + Docker Compose** —— 需要一台有公网 IP 的云服务器。

---

## 方案 A：Render 部署（无需服务器，推荐）

Render 是 PaaS 平台，无需自管服务器，直接从 Git 仓库构建部署。免费层提供 Web 服务（Docker）和 PostgreSQL（90 天试用）。

### 步骤

1. 注册 [Render](https://render.com)（可用 GitHub 账号登录）。
2. 把本项目推送到 GitHub 仓库。
3. 在 Render 控制台：**New → Blueprint**，选择你的仓库，Render 会自动读取 `render.yaml`。
4. Render 自动创建：
   - `traffic-platform`（Web 服务，Docker 构建，托管 API + 前端）
   - `traffic-db`（PostgreSQL 数据库，自动注入 `DATABASE_URL`）
5. 等待构建完成，访问 Render 分配的 `https://xxx.onrender.com` 即可。

> 后端启动时会自动执行 `migrations/*.sql` 建表，无需手动操作。
>
> 免费层注意：Web 服务空闲约 15 分钟会休眠（下次访问需冷启动约几十秒）；免费 PostgreSQL 仅 90 天，到期前建议改用 [Neon](https://neon.tech)（永久免费 Postgres）或 [Supabase](https://supabase.com) 免费库，把连接串填入 `DATABASE_URL` 环境变量即可。

---

## 方案 B：自有云服务器 + Docker Compose

本文档剩余部分说明如何把平台部署到一台有公网 IP 的云服务器。

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
