package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"
)

// Migrate 按文件名顺序执行 migrations/*.sql（幂等建表脚本）。
func (s *PostgresStore) Migrate(ctx context.Context) error {
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("no migration files found in migrations/")
	}

	// 用 simple protocol 执行多语句 SQL 文件（含 CREATE 多条、DO $$ 块）
	cfg := s.pool.Config().ConnConfig.Copy()
	cfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := conn.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("migrate %s: %w", f, err)
		}
	}
	return nil
}
