package migration_runner

import (
	"database/sql"
	"fmt"
	"io/fs"
	"sort"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/file"
)

type Runner struct {
	migrate *migrate.Migrate
}

type Source struct {
	Path     string
	Priority int
	Exclude  []string
	FS       fs.FS
}

func New(sources []Source, connectionString, tableName string) (*Runner, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: tableName,
	})
	if err != nil {
		return nil, fmt.Errorf("create driver: %w", err)
	}

	sourceDriver, err := processSources(sources, connectionString, tableName)
	if err != nil {
		return nil, fmt.Errorf("process sources: %w", err)
	}
	m, err := migrate.NewWithInstance(sources[0].Path, sourceDriver, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("migration runner: %w", err)
	}

	return &Runner{migrate: m}, nil
}

func processSources(sources []Source, connectionString, tableName string) (source.Driver, error) {
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Priority > sources[j].Priority
	})

	var sourceDriver source.Driver
	for _, src := range sources {
		var srcDriver source.Driver
		var err error

		if src.FS != nil {
			srcDriver = NewEmbedSource(src.FS, src.Path)
		} else {
			srcDriver, err = (&file.File{}).Open(src.Path)
			if err != nil {
				return nil, fmt.Errorf("open file source: %w", err)
			}
		}

		filteredSource := &filteredSource{
			Driver:  srcDriver,
			exclude: src.Exclude,
		}

		if sourceDriver == nil {
			sourceDriver = filteredSource
		} else {
			sourceDriver = &chainedSource{
				primary:  sourceDriver,
				fallback: filteredSource,
			}
		}
	}

	return sourceDriver, nil
}

// Up applies all up migrations
func (r *Runner) Up() error {
	return r.migrate.Up()
}

// Down rolls back all or N migrations
func (r *Runner) Down(steps int) error {
	if steps == 0 {
		return r.migrate.Down()
	}
	return r.migrate.Steps(-steps)
}

// Steps applies N up/down steps (positive = up, negative = down)
func (r *Runner) Steps(steps int) error {
	return r.migrate.Steps(steps)
}

// Close releases migration resources
func (r *Runner) Close() {
	r.migrate.Close()
}
