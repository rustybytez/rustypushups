package store

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"

	"github.com/ivansantos/rustypushups/internal/db"
	_ "modernc.org/sqlite"
)

//go:embed migrations.sql
var migrations string

type Store struct {
	*db.Queries
	sqldb *sql.DB
}

func New(dsn string) (*Store, error) {
	sqldb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	sqldb.SetMaxOpenConns(1)
	s := &Store{Queries: db.New(sqldb), sqldb: sqldb}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	stmts := strings.Split(migrations, ";")
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := s.sqldb.ExecContext(context.Background(), stmt); err != nil {
			if !strings.Contains(err.Error(), "already exists") &&
				!strings.Contains(err.Error(), "duplicate column name") {
				return err
			}
		}
	}
	return nil
}
