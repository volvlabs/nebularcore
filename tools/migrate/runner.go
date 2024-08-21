package migrate

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

type Runner struct {
	migrate *migrate.Migrate
}

func NewRunner(migrationsDir, connectionString string) (*Runner, error) {
	migrate, err := migrate.New(migrationsDir, connectionString)
	if err != nil {
		log.Err(err).Msg("unable to start migration")
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
		log.Err(err).Msg("database migration failed")
		return err
	}
	log.Info().Msg("database migration ran successfully")
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
		log.Err(err).Msg("database miration rollback ran failed")
		return err
	}
	log.Info().Msg("database miration rollback ran successfully")
	return nil
}

func (r *Runner) Close() {
	r.migrate.Close()
}
