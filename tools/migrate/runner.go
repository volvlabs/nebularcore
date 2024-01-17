package migrate

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Runner struct {
	migrate *migrate.Migrate
}

func NewRunner(migrationsDir, connectionString string) (*Runner, error) {
	migrate, err := migrate.New(migrationsDir, connectionString)
	if err != nil {
		return nil, err
	}

	r := &Runner{
		migrate: migrate,
	}
	return r, nil
}

func (r *Runner) Run(args ...string) error {
	cmd := "up"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "up":
		return r.up()
	case "down":
		toRevertCount := 0
		if len(args) > 1 {
			toRevertCount, _ = fmt.Sscanf(args[1], "%d", &toRevertCount)
		}
		return r.down(toRevertCount)
	default:
		return fmt.Errorf("command not supported: %q", cmd)
	}
}

func (r *Runner) up() error {
	if err := r.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func (r *Runner) down(toRevertCount int) error {
	var err error
	if toRevertCount == 0 {
		err = r.migrate.Down()
	} else {
		err = r.migrate.Steps(-toRevertCount)
	}

	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func (r *Runner) Close() {
	fmt.Println(r.migrate.Close())
}
