package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/feed"
)

func Identifier(j feed.JobItem) string {
	return normalize(j.Title) + "|" + normalize(j.Company)
}

type JobStore struct {
	path string
	db   *sql.DB
}

func NewJobStore(path string) (*JobStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	db.Exec(`PRAGMA journal_mode = WAL;`)
	db.Exec(`PRAGMA synchronous = NORMAL;`)
	return &JobStore{
		db:   db,
		path: path,
	}, nil
}

func (s *JobStore) Close() error {
	return s.db.Close()
}

func (s *JobStore) createJobsTable(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS jobs (
	hash        INTEGER  PRIMARY KEY,
	source      TEXT     NOT NULL,
	title       TEXT     NOT NULL,
	link        TEXT     NOT NULL,
	company     TEXT     NOT NULL,
	location    TEXT     NOT NULL,

	seniority   TEXT     CHECK (
		seniority IS NULL
		OR seniority IN (%s)
	),
	jobtype     TEXT     CHECK (
		jobtype IS NULL
		OR jobtype IN (%s)
	),

	date        TEXT     NOT NULL,
	
	inserted_at TEXT     DEFAULT CURRENT_TIMESTAMP,
	score       REAL     DEFAULT 1.0
) STRICT;`

	toCheckList := func(l []string) string {
		for i, s := range l {
			l[i] = fmt.Sprintf("'%s'", s)
		}

		return strings.Join(l, ",")
	}

	query := fmt.Sprintf(schema,
		toCheckList(feed.SENIORITIES),
		toCheckList(feed.JOB_TYPES),
	)

	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *JobStore) createTagsTables(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- join table
CREATE TABLE IF NOT EXISTS job_tags (
    job_hash INTEGER NOT NULL REFERENCES jobs(hash) ON DELETE CASCADE,
    tag_id   INTEGER NOT NULL REFERENCES tags(id)   ON DELETE CASCADE,
    PRIMARY KEY (job_hash, tag_id)
);`
	_, err := s.db.ExecContext(ctx, schema)
	return err
}

func (s *JobStore) CreateTables(ctx context.Context) error {
	var err error

	if err = s.createJobsTable(ctx); err != nil {
		return fmt.Errorf("Failed to create Jobs table: %w", err)
	}
	if err = s.createTagsTables(ctx); err != nil {
		return fmt.Errorf("Failed to create Tags tables table: %w", err)
	}

	return nil
}

func (s *JobStore) Insert(
	log *logrus.Entry,
	ctx context.Context,
	j feed.JobItem,
) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	rollback := func(e error) (bool, error) { tx.Rollback(); return false, e }

	const q = `
INSERT OR IGNORE INTO jobs
	(
		hash, source, title, link, company, location,
		seniority, jobType,
		date, score
	)
VALUES
	(
		?,    ?,      ?,     ?,    ?,       ?,
		NULLIF(?, ''), NULLIF(?, ''),
		?,    ?
	);
`

	if j.Seniority == feed.UnknownSeniority {
		log.WithFields(logrus.Fields{
			"seniority": j.Seniority,
			"title":     j.Title,
			"src":       j.Source,
		}).Warn("unknown seniority")
	}

	if j.JobType == feed.UnknownJobType {
		log.WithFields(logrus.Fields{
			"jobType": j.JobType,
			"title":   j.Title,
			"src":     j.Source,
		}).Warn("unknown seniority")
	}

	hash := HashNormalized64(Identifier(j))

	res, err := tx.ExecContext(ctx, q,
		hash,
		j.Source,
		j.Title,
		j.Link,
		j.Company,
		j.Location,

		string(j.Seniority),
		string(j.JobType),
		j.Date.Format(time.RFC3339),
		j.Score,
	)
	if err != nil {
		return rollback(err)
	}

	for _, raw := range j.Tags {
		tagNorm := NormalizeTag(raw)
		if tagNorm == "" {
			continue
		}

		// upsert tag row (ignore if exists)
		if _, err := tx.ExecContext(
			ctx,
			`INSERT OR IGNORE INTO tags(name) VALUES(?)`,
			tagNorm,
		); err != nil {
			return rollback(err)
		}

		var tagID int64
		if err := tx.QueryRowContext(ctx,
			`SELECT id FROM tags WHERE name = ?`, tagNorm).Scan(&tagID); err != nil {
			return rollback(err)
		}

		// link job <-> tag
		if _, err := tx.ExecContext(
			ctx,
			`INSERT OR IGNORE INTO job_tags(job_hash, tag_id) VALUES(?, ?)`,
			hash,
			tagID,
		); err != nil {
			return rollback(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	newRow, err2 := res.RowsAffected()
	if err2 != nil {
		log.WithFields(logrus.Fields{
			"err": err2,
		}).Warn("getting rows affected")
	} else if newRow == 0 {
		log.WithFields(logrus.Fields{
			"identifier": Identifier(j),
		}).Debug("duplicate job skipped")
	}

	return newRow == 1, nil
}

func (s *JobStore) GetJobs(
	log *logrus.Entry,
	ctx context.Context,
	filter string,
) ([]feed.JobItem, error) {
	const q = `
SELECT
	hash, source, title, link, company, location, seniority, jobType, date, score
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

	var out []feed.JobItem
	for rows.Next() {
		var seniorityStr, jobTypeStr, dateStr sql.NullString

		var j feed.JobItem
		if err := rows.Scan(
			&j.Hash,
			&j.Source,
			&j.Title,
			&j.Link,
			&j.Company,
			&j.Location,

			&seniorityStr,
			&jobTypeStr,
			&dateStr,

			&j.Score,
		); err != nil {
			return nil, err
		}

		// NOTE: only parse when present; otherwise leave Unknown*
		if seniorityStr.Valid && seniorityStr.String != "" {
			if j.Seniority, err = feed.ParseSeniority(log, seniorityStr.String); err != nil {
				j.Seniority = feed.UnknownSeniority
			}
		}
		if jobTypeStr.Valid && jobTypeStr.String != "" {
			if j.JobType, err = feed.ParseJobType(log, jobTypeStr.String); err != nil {
				j.JobType = feed.UnknownJobType
			}
		}

		// dates are stored as RFC3339; fall back to yyyy-mm-dd just in case
		if dateStr.Valid && dateStr.String != "" {
			if t, err := time.Parse(time.RFC3339, dateStr.String); err == nil {
				j.Date = t
			} else if t, err := time.Parse("2006-01-02", dateStr.String); err == nil {
				j.Date = t
			} else {
				return nil, fmt.Errorf("invalid date in DB: %q: %w", dateStr.String, err)
			}
		}

		out = append(out, j)
	}
	return out, rows.Err()
}

func (s *JobStore) StreamJobs(
	log *logrus.Entry,
	ctx context.Context,
	filter string,
) (<-chan feed.JobItem, <-chan error) {
	jobs := make(chan feed.JobItem)
	errs := make(chan error, 1)

	go func() {
		defer close(jobs)
		defer close(errs)

		const q = `
SELECT
	hash, source, title, link, company, location, seniority, jobType, date, score
FROM
	jobs
WHERE
	title LIKE ? ORDER BY score DESC, inserted_at DESC;
`
		rows, err := s.db.QueryContext(ctx, q, "%"+filter+"%")
		if err != nil {
			errs <- err
			return
		}
		defer rows.Close()

		for rows.Next() {
			var (
				// FIXME: SqlStringNilwhatever
				seniorityStr string
				jobTypeStr   string
			)
			var j feed.JobItem
			if err := rows.Scan(
				&j.Hash,
				&j.Source,
				&j.Title,
				&j.Link,
				&j.Company,
				&j.Location,
				seniorityStr,
				jobTypeStr,
				&j.Date,
				&j.InsertedAt,
				&j.Score,
			); err != nil {
				errs <- err
				return
			}

			j.Seniority, err = feed.ParseSeniority(log, seniorityStr)
			if err != nil {
				errs <- err
				return
			}

			j.JobType, err = feed.ParseJobType(log, jobTypeStr)
			if err != nil {
				errs <- err
				return
			}

			select {
			case jobs <- j:
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			}
		}

		if err := rows.Err(); err != nil {
			errs <- err
		}
	}()

	return jobs, errs
}

func (s *JobStore) WipeDB() error {
	if err := s.db.Close(); err != nil {
		return err
	}
	return os.Remove(s.path)
}
