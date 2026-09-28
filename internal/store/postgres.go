package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"traffic-platform/internal/model"
)

// PostgresStore 是基于 pgx 连接池的 PostgreSQL 实现。
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore 建立连接池并做连通性检查。
func NewPostgresStore(ctx context.Context, url string) (*PostgresStore, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

// aggTable 将窗口类型映射到聚合表名，避免拼接外部输入造成的注入风险。
func aggTable(w AggWindow) (string, error) {
	switch w {
	case AggWindow1m:
		return "traffic_agg_1m", nil
	case AggWindow5m:
		return "traffic_agg_5m", nil
	default:
		return "", fmt.Errorf("unknown agg window: %s", w)
	}
}

// ---- 原始事件 ----

const insertEventSQL = `
INSERT INTO raw_traffic_events (intersection_id, event_time, vehicle_count, avg_speed, source, created_at)
VALUES ($1, $2, $3, $4, $5, now())
RETURNING id, created_at`

func (s *PostgresStore) InsertEvent(ctx context.Context, e *model.RawTrafficEvent) error {
	return s.pool.QueryRow(ctx, insertEventSQL,
		e.IntersectionID, e.EventTime, e.VehicleCount, e.AvgSpeed, e.Source,
	).Scan(&e.ID, &e.CreatedAt)
}

func (s *PostgresStore) InsertEvents(ctx context.Context, events []*model.RawTrafficEvent) (int, error) {
	if len(events) == 0 {
		return 0, nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // commit 后回滚为无操作

	success := 0
	for _, e := range events {
		if err := tx.QueryRow(ctx, insertEventSQL,
			e.IntersectionID, e.EventTime, e.VehicleCount, e.AvgSpeed, e.Source,
		).Scan(&e.ID, &e.CreatedAt); err != nil {
			continue
		}
		success++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return success, nil
}

const selectEventsBetweenSQL = `
SELECT id, intersection_id, event_time, vehicle_count, avg_speed, source, created_at
FROM raw_traffic_events
WHERE event_time >= $1 AND event_time < $2
ORDER BY event_time`

func (s *PostgresStore) EventsBetween(ctx context.Context, start, end time.Time) ([]*model.RawTrafficEvent, error) {
	rows, err := s.pool.Query(ctx, selectEventsBetweenSQL, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func scanEvents(rows pgx.Rows) ([]*model.RawTrafficEvent, error) {
	var out []*model.RawTrafficEvent
	for rows.Next() {
		var e model.RawTrafficEvent
		if err := rows.Scan(&e.ID, &e.IntersectionID, &e.EventTime, &e.VehicleCount, &e.AvgSpeed, &e.Source, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}
