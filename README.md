# Go 交通数据分析与可视化平台（骨架）

基于 PRD v0.1 实现的全栈骨架（Go 后端 + React 前端），覆盖「数据接入 + 聚合分析 + 看板展示 + 告警」四大模块。骨架阶段仅做可运行的最简结构，不做复杂分析。

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
└── docker-compose.yml              # 本地 PostgreSQL
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
| PATCH | `/api/alerts/:id/resolve` | 处理告警 |
| GET | `/api/admin/import-jobs` | 导入任务状态 |
| POST | `/api/admin/aggregate/run` | 手动触发聚合任务（`window` 为 `1m`/`5m`） |
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

## 定时任务

- 每分钟：聚合上一个 1 分钟窗口到 `traffic_agg_1m`，并评估告警规则
- 每 5 分钟：聚合上一个 5 分钟窗口到 `traffic_agg_5m`

告警规则（骨架最简实现）：

| 规则码 | 级别 | 触发条件 |
|--------|------|----------|
| `flow_spike` | warning | 当前分钟车流高于近 5 分钟均值 1.5 倍 |
| `low_speed` | critical | 平均速度低于 20 km/h |

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
