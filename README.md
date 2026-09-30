# Go 交通数据分析与可视化平台

> 🌐 **线上演示**：http://47.101.32.47

基于 PRD 实现的交通数据分析与可视化平台，覆盖「数据接入 → 聚合分析 → 告警 → 看板展示」完整链路。采用 Go 后端 + React 前端 + PostgreSQL，已部署上线。

## 功能特性

**后端**

- 数据接入：单条事件写入、CSV / JSON 批量导入、模拟数据生成
- 聚合分析：1 分钟 / 5 分钟时间窗口聚合，计算车流量、平均速度、拥堵指数
- 告警引擎：流量突增、连续低速两条规则，支持「新建 → 已确认 → 已处理」状态流转
- 定时任务：robfig/cron 每分钟聚合并评估告警
- 自动建表迁移、统一响应结构、错误码规范

**前端**

- 6 个页面：总览看板、趋势分析、路口排行、告警、数据导入、任务管理
- ECharts 多系列趋势图、拥堵排行柱状图、彩色徽章告警列表
- 时间范围切换、级别/状态筛选、模拟数据一键生成

## 技术栈

**后端**

- Web 框架：Gin
- 数据库：PostgreSQL（`pgx/v5` 连接池）
- 定时任务：robfig/cron/v3
- 配置：环境变量

**前端**

- 构建：Vite + React 18 + TypeScript
- 路由：React Router v6
- 图表：ECharts
- HTTP：axios

### 系统架构

```mermaid
flowchart LR
    SOURCE["数据源 / 模拟器"] --> API["Go API"]
    API --> RAW["原始数据表 raw_traffic_events"]
    RAW --> AGG["聚合任务 cron 1m/5m"]
    AGG --> ALERT["告警规则"]
    AGG --> DASH["Dashboard API"]
    ALERT --> DASH
    DASH --> WEB["React 看板前端"]
```

## 目录结构

```
.
├── cmd/server/main.go              # 后端入口：装配 + 启动 cron + 优雅退出
├── internal/                       # 后端核心代码
│   ├── config/config.go            # 配置（数据库连接串、聚合阈值等）
│   ├── model/model.go              # 数据模型
│   ├── store/                      # 数据访问层（接口 + PostgreSQL 实现）
│   ├── service/                    # 业务逻辑（接入/聚合/告警/看板）
│   ├── aggregation/scheduler.go    # 定时聚合与告警调度
│   ├── handler/                    # HTTP handler + 统一响应
│   └── router/router.go            # 路由注册
├── frontend/                       # 前端（Vite + React + ECharts）
│   ├── src/pages/                  # 6 个页面
│   ├── src/components/             # 布局、指标卡片、图表封装
│   └── src/api/                    # API 调用封装
├── migrations/                   # 建表脚本（001 初始化 + 002 增量迁移）
├── deploy/                         # 生产部署（Dockerfile + Nginx + 部署文档）
├── scripts/e2e_test.ps1            # 端到端测试脚本
├── docker-compose.yml              # 本地开发（PostgreSQL）
└── docker-compose.prod.yml         # 生产部署编排
```

## 快速开始

### 1. 启动 PostgreSQL

```bash
docker-compose up -d
```

首次启动会自动执行 `migrations/001_init.sql` 建表。

### 2. 拉取依赖并运行

```bash
go mod tidy
go run ./cmd/server
```

服务默认监听 `:8080`。若数据库不是默认连接串，通过环境变量指定：

```bash
$env:DATABASE_URL = "postgres://traffic:traffic@localhost:5432/traffic?sslmode=disable"
go run ./cmd/server
```

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端默认运行在 `http://localhost:5173`，开发服务器会把 `/api` 请求代理到后端 `:8080`。

> 生产构建：`cd frontend && npm run build`（产物输出到 `frontend/dist`）

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SERVER_ADDR` | `:8080` | HTTP 监听地址 |
| `DATABASE_URL` | `postgres://traffic:traffic@localhost:5432/traffic?sslmode=disable` | PG 连接串 |
| `FREE_FLOW_SPEED` | `60.0` | 拥堵指数自由流速度（km/h） |
| `LOW_SPEED_THRESHOLD` | `20.0` | 低速告警阈值（km/h） |
| `FLOW_SPIKE_FACTOR` | `1.5` | 流量突增告警倍数阈值 |
| `LOW_SPEED_CONSECUTIVE` | `3` | 连续低速告警需连续低于阈值的窗口数 |

## 接口清单

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/traffic/events` | 写入单条交通事件 |
| POST | `/api/traffic/import` | CSV/批量导入（JSON 数组或 multipart CSV） |
| POST | `/api/traffic/simulate` | 生成模拟交通事件（演示/测试入口） |
| GET | `/api/dashboard/overview` | 概览卡片数据 |
| GET | `/api/dashboard/trend` | 趋势图数据（支持 `start`/`end`） |
| GET | `/api/dashboard/intersections/top` | 拥堵路口排行（支持 `limit`） |
| GET | `/api/alerts` | 告警列表（支持 `level`/`status` 筛选） |
| PATCH | `/api/alerts/:id/ack` | 确认告警（新建 → 已确认） |
| PATCH | `/api/alerts/:id/resolve` | 处理告警（→ 已处理） |
| GET | `/api/admin/import-jobs` | 导入任务状态 |
| POST | `/api/admin/aggregate/run` | 手动触发聚合（`window` 为 `1m`/`5m`，可选 `minutes` 回填历史窗口） |
| GET | `/healthz` | 健康检查 |

### 请求示例

写入单条事件：

```bash
curl -X POST http://localhost:8080/api/traffic/events \
  -H "Content-Type: application/json" \
  -d '{
    "intersectionId": "A-101",
    "timestamp": "2026-04-01T08:30:00+08:00",
    "vehicleCount": 42,
    "avgSpeed": 18.6,
    "source": "simulator"
  }'
```

批量导入 JSON 数组：

```bash
curl -X POST http://localhost:8080/api/traffic/import \
  -H "Content-Type: application/json" \
  -d '[{"intersectionId":"A-101","timestamp":"2026-04-01T08:30:00+08:00","vehicleCount":42,"avgSpeed":18.6,"source":"simulator"}]'
```

导入 CSV（列顺序：`intersection_id,event_time,vehicle_count,avg_speed,source`）：

```bash
curl -X POST http://localhost:8080/api/traffic/import \
  -F "file=@events.csv"
```

查询告警：

```bash
curl "http://localhost:8080/api/alerts?level=critical&status=new"
```

处理告警：

```bash
curl -X PATCH http://localhost:8080/api/alerts/1/resolve
```

## 快速演示

部署后一条命令生成演示数据并回填聚合，看板即可展示丰富趋势：

```bash
# 生成 20 个路口 × 60 分钟 × 每分钟 5 条 = 6000 条模拟数据
curl -X POST http://localhost:8080/api/traffic/simulate \
  -H "Content-Type: application/json" \
  -d '{"intersections":20,"minutes":60,"eventsPerMinute":5}'

# 回填聚合最近 60 分钟（趋势图 60 个点、排行 20 个路口）
curl -X POST http://localhost:8080/api/admin/aggregate/run \
  -H "Content-Type: application/json" \
  -d '{"window":"1m","minutes":60}'

curl -X POST http://localhost:8080/api/admin/aggregate/run \
  -H "Content-Type: application/json" \
  -d '{"window":"5m","minutes":60}'
```

也可以直接在前端「数据导入」页点「生成模拟数据」按钮。

## 定时任务

- 每分钟：聚合上一个 1 分钟窗口到 `traffic_agg_1m`，并评估告警规则
- 每 5 分钟：聚合上一个 5 分钟窗口到 `traffic_agg_5m`

告警规则（骨架最简实现）：

| 规则码 | 级别 | 触发条件 |
|--------|------|----------|
| `flow_spike` | warning | 当前分钟车流高于近 5 分钟均值 1.5 倍 |
| `low_speed` | critical | 平均速度连续 N 个窗口低于阈值（默认连续 3 分钟 < 20 km/h） |

## 数据模型

对齐 PRD 第 6 节：`raw_traffic_events`、`traffic_agg_1m`、`traffic_agg_5m`、`alerts`、`import_jobs`，详见 `migrations/001_init.sql`。

## 统一响应结构

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

`code` 为 0 表示成功，非 0 为业务错误码（400xx 客户端错误、500xx 服务端错误）。

## 端到端测试

验证两条核心链路：

- 场景 1：接入数据 → 聚合任务 → 看板更新
- 场景 2：触发告警条件 → 告警记录生成

```bash
# 1. 启动 PostgreSQL（如未启动）
docker-compose up -d

# 2. 运行端到端测试（自动启动后端 + 测试 + 清理）
./scripts/e2e_test.ps1

# 后端已手动启动时
./scripts/e2e_test.ps1 -SkipBackend
```

脚本会自动断言每个步骤（生成数据、聚合、看板数据、告警生成、告警处理），输出 `[PASS]`/`[FAIL]`。

## 生产部署

部署到公共网络（云服务器 + Docker Compose）的完整方案见 [deploy/DEPLOY.md](deploy/DEPLOY.md)：

```bash
docker compose -f docker-compose.prod.yml up -d --build
```
