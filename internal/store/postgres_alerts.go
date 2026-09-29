package store

import (
	"context"

	"github.com/jackc/pgx/v5"

	"traffic-platform/internal/model"
)

// ---- 告警 ----

const insertAlertSQL = `
INSERT INTO alerts (intersection_id, level, rule_code, status, message, created_at)
VALUES ($1, $2, $3, $4, $5, now())
RETURNING id, created_at`

func (s *PostgresStore) InsertAlert(ctx context.Context, a *model.Alert) error {
	return s.pool.QueryRow(ctx, insertAlertSQL,
		a.IntersectionID, a.Level, a.RuleCode, a.Status, a.Message,
	).Scan(&a.ID, &a.CreatedAt)
}

func (s *PostgresStore) ListAlerts(ctx context.Context, level, status string) ([]*model.Alert, error) {
	const sql = `
SELECT id, intersection_id, level, rule_code, status, message, created_at, resolved_at
FROM alerts
WHERE ($1 = '' OR level = $1) AND ($2 = '' OR status = $2)
ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, sql, level, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*model.Alert
	for rows.Next() {
		var a model.Alert
		if err := rows.Scan(&a.ID, &a.IntersectionID, &a.Level, &a.RuleCode, &a.Status, &a.Message, &a.CreatedAt, &a.ResolvedAt); err != nil {
			return nil, err
		}
		out = append(out, &a)
	}
	return out, rows.Err()
}

func (s *PostgresStore) ResolveAlert(ctx context.Context, id int64) error {
	const sql = `
UPDATE alerts
SET status = 'resolved', resolved_at = now()
WHERE id = $1`
	ct, err := s.pool.Exec(ctx, sql, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *PostgresStore) HasActiveAlert(ctx context.Context, intersectionID, ruleCode string) (bool, error) {
	const sql = `
SELECT EXISTS(
    SELECT 1 FROM alerts
    WHERE intersection_id = $1 AND rule_code = $2 AND status <> 'resolved'
)`
	var exists bool
	if err := s.pool.QueryRow(ctx, sql, intersectionID, ruleCode).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// ---- 导入任务 ----

const insertImportJobSQL = `
INSERT INTO import_jobs (filename, status, total_rows, success_rows, failed_rows, error_details, created_at)
VALUES ($1, $2, $3, $4, $5, $6, now())
RETURNING id, created_at`

func (s *PostgresStore) InsertImportJob(ctx context.Context, j *model.ImportJob) error {
	return s.pool.QueryRow(ctx, insertImportJobSQL,
		j.Filename, j.Status, j.TotalRows, j.SuccessRows, j.FailedRows, j.ErrorDetails,
	).Scan(&j.ID, &j.CreatedAt)
}

func (s *PostgresStore) ListImportJobs(ctx context.Context) ([]*model.ImportJob, error) {
	const sql = `
SELECT id, filename, status, total_rows, success_rows, failed_rows, error_details, created_at
FROM import_jobs
ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*model.ImportJob
	for rows.Next() {
		var j model.ImportJob
		if err := rows.Scan(&j.ID, &j.Filename, &j.Status, &j.TotalRows, &j.SuccessRows, &j.FailedRows, &j.ErrorDetails, &j.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &j)
	}
	return out, rows.Err()
}
