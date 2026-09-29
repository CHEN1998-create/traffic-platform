package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"traffic-platform/internal/model"
)

const aggColumns = `id, intersection_id, window_start, total_vehicles, avg_speed, congestion_index`

func (s *PostgresStore) InsertAgg(ctx context.Context, window AggWindow, a *model.TrafficAgg) error {
	table, err := aggTable(window)
	if err != nil {
		return err
	}
	// UPSERT：同一 (intersection_id, window_start) 重复聚合时覆盖而非重复插入，
	// 保证聚合任务可重复执行且结果稳定（PRD 第 9 节）。
	sql := fmt.Sprintf(`
INSERT INTO %s (intersection_id, window_start, total_vehicles, avg_speed, congestion_index)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (intersection_id, window_start)
DO UPDATE SET total_vehicles = EXCLUDED.total_vehicles,
              avg_speed = EXCLUDED.avg_speed,
              congestion_index = EXCLUDED.congestion_index
RETURNING id`, table)
	return s.pool.QueryRow(ctx, sql,
		a.IntersectionID, a.WindowStart, a.TotalVehicles, a.AvgSpeed, a.CongestionIndex,
	).Scan(&a.ID)
}

func (s *PostgresStore) LatestAgg(ctx context.Context, window AggWindow) ([]*model.TrafficAgg, error) {
	table, err := aggTable(window)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`
SELECT %s FROM %s
WHERE window_start = (SELECT max(window_start) FROM %s)
ORDER BY congestion_index DESC`, aggColumns, table, table)
	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAggs(rows)
}

func (s *PostgresStore) AggTrend(ctx context.Context, window AggWindow, start, end time.Time) ([]*model.TrafficAgg, error) {
	table, err := aggTable(window)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`
SELECT min(id), '' AS intersection_id, window_start,
       sum(total_vehicles) AS total_vehicles,
       CASE WHEN sum(total_vehicles) > 0
            THEN sum(total_vehicles * avg_speed) / sum(total_vehicles)
            ELSE 0 END AS avg_speed,
       CASE WHEN sum(total_vehicles) > 0
            THEN sum(total_vehicles * congestion_index) / sum(total_vehicles)
            ELSE 0 END AS congestion_index
FROM %s
WHERE window_start >= $1 AND window_start < $2
GROUP BY window_start
ORDER BY window_start`, table)
	rows, err := s.pool.Query(ctx, sql, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAggs(rows)
}

func (s *PostgresStore) AggRecentForIntersection(ctx context.Context, window AggWindow, intersectionID string, since time.Time) ([]*model.TrafficAgg, error) {
	table, err := aggTable(window)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`
SELECT %s FROM %s
WHERE intersection_id = $1 AND window_start >= $2
ORDER BY window_start DESC`, aggColumns, table)
	rows, err := s.pool.Query(ctx, sql, intersectionID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAggs(rows)
}

func (s *PostgresStore) TopIntersections(ctx context.Context, window AggWindow, limit int) ([]*model.TrafficAgg, error) {
	table, err := aggTable(window)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`
SELECT min(id), intersection_id, max(window_start) AS window_start,
       sum(total_vehicles) AS total_vehicles,
       CASE WHEN sum(total_vehicles) > 0
            THEN sum(total_vehicles * avg_speed) / sum(total_vehicles)
            ELSE 0 END AS avg_speed,
       CASE WHEN sum(total_vehicles) > 0
            THEN sum(total_vehicles * congestion_index) / sum(total_vehicles)
            ELSE 0 END AS congestion_index
FROM %s
WHERE window_start >= now() - interval '30 minutes'
GROUP BY intersection_id
ORDER BY congestion_index DESC
LIMIT $1`, table)
	rows, err := s.pool.Query(ctx, sql, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAggs(rows)
}

func scanAggs(rows pgx.Rows) ([]*model.TrafficAgg, error) {
	var out []*model.TrafficAgg
	for rows.Next() {
		var a model.TrafficAgg
		if err := rows.Scan(&a.ID, &a.IntersectionID, &a.WindowStart, &a.TotalVehicles, &a.AvgSpeed, &a.CongestionIndex); err != nil {
			return nil, err
		}
		out = append(out, &a)
	}
	return out, rows.Err()
}
