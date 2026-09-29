-- =============================================================
-- 增量迁移：为已建库的环境补齐唯一约束与错误详情字段
-- （全新环境请直接使用 001_init.sql，本文件用于增量升级）
-- =============================================================

-- 为 1m 聚合表补唯一约束（保证聚合幂等）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_agg1m_intersection_window'
    ) THEN
        ALTER TABLE traffic_agg_1m
            ADD CONSTRAINT uq_agg1m_intersection_window UNIQUE (intersection_id, window_start);
    END IF;
END $$;

-- 为 5m 聚合表补唯一约束
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_agg5m_intersection_window'
    ) THEN
        ALTER TABLE traffic_agg_5m
            ADD CONSTRAINT uq_agg5m_intersection_window UNIQUE (intersection_id, window_start);
    END IF;
END $$;

-- import_jobs 增加错误详情列
ALTER TABLE import_jobs ADD COLUMN IF NOT EXISTS error_details text;
