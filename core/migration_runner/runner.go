package migration_runner

import (
	"database/sql"
	"fmt"
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
}

func New(sources []Source, connectionString, tableName string) (*Runner, error) {
	// Create database driver
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Create database instance
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: tableName,
	})
	if err != nil {
		return nil, fmt.Errorf("create driver: %w", err)
	}

	// Process migration sources
	sourceDriver, err := processSources(sources)
	if err != nil {
		return nil, fmt.Errorf("process sources: %w", err)
	}

	// Create migrate instance with first source path as identifier
	m, err := migrate.NewWithInstance(sources[0].Path, sourceDriver, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("migration runner: %w", err)
	}

	return &Runner{migrate: m}, nil
}

func processSources(sources []Source) (source.Driver, error) {
	// Sort sources by priority (highest first)
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Priority > sources[j].Priority
	})

	// Chain sources together
	var sourceDriver source.Driver
	for _, src := range sources {
		// Create file source
		fileSource, err := (&file.File{}).Open(src.Path)
		if err != nil {
			return nil, fmt.Errorf("open file source: %w", err)
		}

		// Filter out excluded files
		filteredSource := &filteredSource{
			Driver:  fileSource,
			exclude: src.Exclude,
		}

		// Chain sources
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
