package migrate

import (
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

type Runner struct {
	migrate           *migrate.Migrate
	migrationsDir     string
	connectionString  string
	schemaSearchParam string
}

func NewRunner(migrationsDir, connectionString string) (*Runner, error) {
	migrate, err := migrate.New(migrationsDir, connectionString)
	if err != nil {
		log.Err(err).Msg("unable to start migration")
		return nil, err
	}

	r := &Runner{
		migrate:          migrate,
		migrationsDir:    migrationsDir,
		connectionString: connectionString,
	}
	return r, nil
}

// WithSchema creates a new Runner instance with the specified schema
func (r *Runner) WithSchema(schemaName string) (*Runner, error) {
	connectionString := r.connectionString
	if !strings.Contains(connectionString, "search_path=") {
		if strings.Contains(connectionString, "?") {
			connectionString = fmt.Sprintf("%s&search_path=%s", connectionString, schemaName)
		} else {
			connectionString = fmt.Sprintf("%s?search_path=%s", connectionString, schemaName)
		}
	} else {
		parts := strings.Split(connectionString, "search_path=")
		prefix := parts[0]
		suffix := ""
		if len(parts) > 1 && strings.Contains(parts[1], "&") {
			suffixParts := strings.SplitN(parts[1], "&", 2)
			suffix = "&" + suffixParts[1]
		}
		connectionString = fmt.Sprintf("%ssearch_path=%s%s", prefix, schemaName, suffix)
	}

	newMigrate, err := migrate.New(r.migrationsDir, connectionString)
	if err != nil {
		log.Err(err).Msgf("unable to start migration for schema %s", schemaName)
		return nil, err
	}

	return &Runner{
		migrate:           newMigrate,
		migrationsDir:     r.migrationsDir,
		connectionString:  connectionString,
		schemaSearchParam: schemaName,
	}, nil
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
			if _, err := fmt.Sscanf(args[1], "%d", &toRevertCount); err != nil {
				return fmt.Errorf("invalid count %q: must be an integer", args[1])
			}
		}
		return r.down(toRevertCount)
	case "goto":
		if len(args) < 2 {
			return fmt.Errorf("goto requires a target version number")
		}
		var version uint
		if _, err := fmt.Sscanf(args[1], "%d", &version); err != nil {
			return fmt.Errorf("invalid version %q: must be a positive integer", args[1])
		}
		return r.goTo(version)
	default:
		return fmt.Errorf("command not supported: %q", cmd)
	}
}

func (r *Runner) up() error {
	schemaInfo := ""
	if r.schemaSearchParam != "" {
		schemaInfo = fmt.Sprintf(" for schema '%s'", r.schemaSearchParam)
	}

	if err := r.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		log.Err(err).Msgf("database migration failed%s", schemaInfo)
		return err
	}
	log.Info().Msgf("database migration ran successfully%s", schemaInfo)
	return nil
}

func (r *Runner) down(toRevertCount int) error {
	schemaInfo := ""
	if r.schemaSearchParam != "" {
		schemaInfo = fmt.Sprintf(" for schema '%s'", r.schemaSearchParam)
	}

	var err error
	if toRevertCount == 0 {
		err = r.migrate.Down()
	} else {
		err = r.migrate.Steps(-toRevertCount)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Err(err).Msgf("database migration rollback failed%s", schemaInfo)
		return err
	}
	log.Info().Msgf("database migration rollback ran successfully%s", schemaInfo)
	return nil
}

func (r *Runner) goTo(version uint) error {
	schemaInfo := ""
	if r.schemaSearchParam != "" {
		schemaInfo = fmt.Sprintf(" for schema '%s'", r.schemaSearchParam)
	}

	if err := r.migrate.Migrate(version); err != nil && err != migrate.ErrNoChange {
		log.Err(err).Msgf("database migration goto version %d failed%s", version, schemaInfo)
		return err
	}
	log.Info().Msgf("database migration moved to version %d successfully%s", version, schemaInfo)
	return nil
}

func (r *Runner) Close() {
	r.migrate.Close()
}
