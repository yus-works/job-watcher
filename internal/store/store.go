package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yus-works/job-watcher/internal/feed"
	"github.com/yus-works/job-watcher/internal/logging"
)

type Job struct {
	ID       string
	Hash     int64
	Source   string
	Title    string
	Link     string
	Company  string
	Location string

	Seniority feed.Seniority
	JobType   feed.JobType
	Date      time.Time

	InsertedAt time.Time
	Score      float64
}

func Identifier(j Job) string {
	return j.Title + "|" + j.Company
}

func FromFeedItem(fi feed.Item) Job {
	return Job{}
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

func (s *JobStore) CreateTables(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS jobs (
	id          TEXT     PRIMARY KEY,
	hash        INTEGER  NOT NULL UNIQUE,
	source      TEXT     NOT NULL,
	title       TEXT     NOT NULL,
	link        TEXT     NOT NULL,
	company     TEXT     NOT NULL,
	location    TEXT     NOT NULL,

	seniority   TEXT     NOT NULL CHECK (seniority IN (%s)),
	jobtype     TEXT     NOT NULL CHECK (jobtype IN (%s)),
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

func (s *JobStore) Insert(ctx context.Context, j Job) (bool, error) {
	// insertedAt skipped bcs db default
	const q = `
INSERT OR IGNORE INTO jobs
	(id, hash, source, title, link, company, location, seniority, jobType, date, score)
VALUES
	(?,  ?,    ?,      ?,     ?,    ?,       ?,        ?,         ?,       ?,    ?);`

	hash := HashNormalized64(Identifier(j))

	res, err := s.db.ExecContext(
		ctx, q,

		j.ID,
		hash,
		j.Source,
		j.Title,
		j.Link,
		j.Company,
		j.Location,

		string(j.Seniority),
		string(j.JobType),
		j.Date.Format(time.RFC3339),

		// insertedAt skipped bcs db default
		j.Score,
	)

	if err != nil {
		log.Println("ERROR: ", err)
		return err
	}

	inserted, err2 := res.RowsAffected()
	if err2 != nil {
		logging.From(ctx).Warn("Getting rows affected: %v", err2)
	} else if inserted == 0 {
		logging.From(ctx).Debug("duplicate job skipped: %s", Identifier(j))
	}

	return !inserted, nil
}

func (s *JobStore) GetJobs(ctx context.Context, filter string) ([]Job, error) {
	const q = `
SELECT
	id, hash, source, title, link, company, location, seniority, jobType, date, score
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
		var (
			seniorityStr string
			jobTypeStr   string
		)
		var j Job
		if err := rows.Scan(
			&j.ID,
			&j.Hash,
			&j.Source,
			&j.Title,
			&j.Link,
			&j.Company,
			&j.Location,

			&j.Date,

			&j.InsertedAt,
			&j.Score,
		); err != nil {
			return nil, err
		}

		if j.Seniority, err = feed.ParseSeniority(seniorityStr); err != nil {
			return nil, err
		}
		if j.JobType, err = feed.ParseJobType(jobTypeStr); err != nil {
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
