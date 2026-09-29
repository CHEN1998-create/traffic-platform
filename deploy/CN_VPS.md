# 国内云服务器部署指南（阿里云 / 腾讯云轻量）

适用：用微信/支付宝付款，把交通数据分析平台部署到国内公网。

## 一、购买服务器

### 购买建议

| 项 | 建议 |
|----|------|
| 平台 | 阿里云「轻量应用服务器」或 腾讯云「轻量应用服务器」 |
| 配置 | **2核2G** 起步（跑 Docker + PostgreSQL + 后端 + Nginx 足够） |
| 带宽 | 3M 及以上 |
| 系统镜像 | **Ubuntu 22.04**（或 Debian 12） |
| 地域 | 国内地域（华北/华东，离你近） |
| 价格 | 新用户活动价约 **50~100 元/年** |

### 购买入口

- 阿里云：控制台搜索「轻量应用服务器」，新用户有优惠价
- 腾讯云：https://cloud.tencent.com/product/lighthouse

购买时注意选择 **Ubuntu 22.04 系统镜像**（Docker 支持最好），付款用微信/支付宝。

## 二、登录服务器

购买成功后，在控制台能看到**公网 IP**。登录方式二选一：

1. **控制台「远程连接」**：阿里云/腾讯云控制台都提供网页版终端，直接点开即可。
2. **本地 SSH**（Windows PowerShell / 终端）：
   ```bash
   ssh root@你的公网IP
   ```
   默认用户是 `root`（轻量服务器），密码在购买时设置或控制台重置。

## 三、安装 Docker

登录后执行（Ubuntu/Debian 通用）：

```bash
# 一键安装 Docker + Compose 插件
curl -fsSL https://get.docker.com | sh

# 启动并设为开机自启
systemctl enable docker && systemctl start docker

# 验证
docker --version
docker compose version
```

> 若 `get.docker.com` 下载慢，可先配置国内镜像源，或改用：
> `curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | apt-key add -`（阿里云源，略复杂）。

## 四、上传代码

### 方式一：git clone（推荐）

```bash
git clone https://github.com/CHEN1998-create/traffic-platform.git
cd traffic-platform
```

> 国内服务器 clone GitHub 可能较慢，可改用镜像：
> `git clone https://ghfast.top/https://github.com/CHEN1998-create/traffic-platform.git`

### 方式二：本地打包上传（无 git 时）

```bash
# 本地 Windows（PowerShell）：打包（排除 node_modules/dist/.git）
tar -czf traffic-platform.tar.gz --exclude=frontend/node_modules --exclude=frontend/dist --exclude=.git .

# 上传到服务器
scp traffic-platform.tar.gz root@你的公网IP:/root/

# 服务器上解压
cd /root && tar -xzf traffic-platform.tar.gz
```

## 五、部署

```bash
cd traffic-platform

# 设置数据库强密码（生产必须改，不要用默认 traffic）
export POSTGRES_PASSWORD="你的强密码"

# 构建并启动（首次构建需几分钟）
docker compose -f docker-compose.prod.yml up -d --build
```

## 六、放行端口（防火墙）

阿里云/腾讯云轻量服务器在 **控制台 → 防火墙** 里添加规则：

- 放行 **80**（HTTP）
- （如需 HTTPS）放行 **443**

> 轻量服务器用的是「防火墙」而不是「安全组」，位置在控制台对应实例的「防火墙」标签页。

## 七、验证

```bash
# 服务器内验证后端
curl http://localhost/healthz
# 期望返回 {"code":0,"message":"ok","data":{"status":"ok"}}

# 浏览器访问（公网）
# http://你的公网IP
```

打开公网 IP 应能看到「交通数据分析平台」看板页面。

## 八、常用运维命令

```bash
docker compose -f docker-compose.prod.yml ps            # 查看状态
docker compose -f docker-compose.prod.yml logs -f backend   # 看后端日志
docker compose -f docker-compose.prod.yml restart      # 重启
docker compose -f docker-compose.prod.yml down         # 停止
```

## 常见问题

| 问题 | 处理 |
|------|------|
| 后端起不来（连不上数据库） | `docker compose logs backend` 查日志，确认 `POSTGRES_PASSWORD` 一致 |
| 首次建表未执行 | 挂载的 `001_init.sql` 只在**数据卷为空**时执行；已建过则手动跑 `002_aggregation_idempotency.sql` |
| 公网访问 404 | 检查防火墙是否放行 80 端口 |
| 想用 HTTPS | 需先准备域名，再申请证书（见 `deploy/DEPLOY.md` 方案 B） |

## 生产环境建议

- `POSTGRES_PASSWORD` 用强随机密码（`openssl rand -hex 16` 生成）。
- 定期备份数据库：`docker exec $(docker compose -f docker-compose.prod.yml ps -q postgres) pg_dump -U traffic traffic > backup.sql`
