package store

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Job struct {
	ID         string
	Title      string
	URL        string
	Company    string
	InsertedAt time.Time
	Score      float64
}

type JobStore struct {
	path string
	db *sql.DB
}

func NewJobStore(path string) (*JobStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	db.Exec(`PRAGMA journal_mode = WAL;`)
	db.Exec(`PRAGMA synchronous = NORMAL;`)
	return &JobStore{
		db: db,
		path: path,
	}, nil
}

func (s *JobStore) Close() error {
	return s.db.Close()
}

func (s *JobStore) CreateTables(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS jobs (
	id          TEXT     PRIMARY KEY,
	title       TEXT     NOT NULL,
	url         TEXT     NOT NULL,
	company     TEXT     NOT NULL,
	inserted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	score       REAL     DEFAULT 1.0
);`

	_, err := s.db.ExecContext(ctx, schema)
	return err
}

func (s *JobStore) Insert(ctx context.Context, j Job) error {
	const q = `
INSERT OR IGNORE INTO
	jobs(id,title,url,company)
VALUES(?,?,?,?);`

	res, err := s.db.ExecContext(ctx, q, j.ID, j.Title, j.URL, j.Company)
	if err != nil {
		log.Println("ERROR: ", err)
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		log.Printf("duplicate %s", j.ID)
	}

	return nil
}

func (s *JobStore) GetJobs(ctx context.Context, filter string) ([]Job, error) {
	const q = `
SELECT
	id, title, url, inserted_at, score, company
FROM
	jobs
WHERE
	title LIKE ? ORDER BY score DESC, inserted_at DESC;
`
	rows, err := s.db.QueryContext(ctx, q, "%"+filter+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(
			&j.ID, &j.Title, &j.URL, &j.InsertedAt, &j.Score, &j.Company,
		); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (s *JobStore) WipeDB() error {
	if err := s.db.Close(); err != nil {
		return err
	}
	return os.Remove(s.path)
}
