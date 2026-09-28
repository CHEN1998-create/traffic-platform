-- =============================================================
-- Go 交通数据分析平台 - 初始化建表脚本
-- 对齐 PRD 第 6 节数据表定义
-- =============================================================

-- 1. 原始交通事件表
CREATE TABLE IF NOT EXISTS raw_traffic_events (
    id              bigserial PRIMARY KEY,
    intersection_id text        NOT NULL,
    event_time      timestamptz NOT NULL,
    vehicle_count   int         NOT NULL,
    avg_speed       numeric     NOT NULL,
    source          text,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_raw_events_time         ON raw_traffic_events (event_time);
CREATE INDEX IF NOT EXISTS idx_raw_events_intersection ON raw_traffic_events (intersection_id);

-- 2. 1 分钟聚合表
CREATE TABLE IF NOT EXISTS traffic_agg_1m (
    id               bigserial PRIMARY KEY,
    intersection_id  text        NOT NULL,
    window_start     timestamptz NOT NULL,
    total_vehicles   int         NOT NULL,
    avg_speed        numeric     NOT NULL,
    congestion_index numeric     NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_agg1m_window       ON traffic_agg_1m (window_start);
CREATE INDEX IF NOT EXISTS idx_agg1m_intersection ON traffic_agg_1m (intersection_id);

-- 3. 5 分钟聚合表
CREATE TABLE IF NOT EXISTS traffic_agg_5m (
    id               bigserial PRIMARY KEY,
    intersection_id  text        NOT NULL,
    window_start     timestamptz NOT NULL,
    total_vehicles   int         NOT NULL,
    avg_speed        numeric     NOT NULL,
    congestion_index numeric     NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_agg5m_window       ON traffic_agg_5m (window_start);
CREATE INDEX IF NOT EXISTS idx_agg5m_intersection ON traffic_agg_5m (intersection_id);

-- 4. 告警表
CREATE TABLE IF NOT EXISTS alerts (
    id              bigserial PRIMARY KEY,
    intersection_id text        NOT NULL,
    level           text        NOT NULL,
    rule_code       text        NOT NULL,
    status          text        NOT NULL DEFAULT 'new',
    message         text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    resolved_at     timestamptz
);

CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts (status);
CREATE INDEX IF NOT EXISTS idx_alerts_rule   ON alerts (rule_code);

-- 5. 导入任务表
CREATE TABLE IF NOT EXISTS import_jobs (
    id           bigserial PRIMARY KEY,
    filename     text        NOT NULL,
    status       text        NOT NULL,
    total_rows   int         NOT NULL DEFAULT 0,
    success_rows int         NOT NULL DEFAULT 0,
    failed_rows  int         NOT NULL DEFAULT 0,
    created_at   timestamptz NOT NULL DEFAULT now()
);
