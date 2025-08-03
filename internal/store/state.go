package store

import (
	"context"
	"database/sql"
	"time"
)

const lastRefreshKey = "last_refresh_at"

func (s *JobStore) SetLastRefresh(ctx context.Context, t time.Time) error {
	q := `
INSERT INTO meta(key, value, updated_at)
VALUES(?, ?, ?)
ON CONFLICT(key) DO UPDATE SET
	value = excluded.value,
	updated_at = excluded.updated_at
`

	_, err := s.db.ExecContext(
		ctx,
		q,
		lastRefreshKey,
		t.UTC().Format(time.RFC3339Nano),
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *JobStore) LastRefresh(ctx context.Context) (time.Time, error) {
	var raw string

	q := `SELECT value FROM meta WHERE key = ?`

	err := s.db.QueryRowContext(ctx, q, lastRefreshKey).Scan(&raw)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}

	// try RFC3339Nano then RFC3339
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, raw)
}
