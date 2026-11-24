package main

import (
	"errors"
	"flag"
	"fmt"
	"llm-wiki-server/internal/config"
	"log"
	"net/url"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jinzhu/configor"
)

const (
	defaultConfigPath     = "config.yml"
	defaultMigrationsPath = "migrations"
)

type options struct {
	configPath     string
	migrationsPath string
	direction      string
	steps          int
	version        int
}

func main() {
	log.SetOutput(os.Stdout)

	opts := parseFlags()
	if err := opts.validate(); err != nil {
		log.Fatalf("invalid options: %v", err)
	}

	var cfg config.Configs
	if err := configor.Load(&cfg, opts.configPath); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	m, err := newMigrator(cfg.DB, opts.migrationsPath)
	if err != nil {
		log.Fatalf("failed to init migrator: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("failed to close migration source: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("failed to close migration database: %v", dbErr)
		}
	}()

	if err := applyMigration(m, opts); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

func parseFlags() options {
	var opts options
	flag.StringVar(&opts.configPath, "config", defaultConfigPath, "path to the config file")
	flag.StringVar(&opts.migrationsPath, "path", defaultMigrationsPath, "path to the migrations directory")
	flag.StringVar(&opts.direction, "direction", "up", "migration direction: up, down, version, force, goto")
	flag.IntVar(&opts.steps, "steps", 0, "number of migrations to apply for up/down (0 means all)")
	flag.IntVar(&opts.version, "version", -1, "target version for force/goto")
	flag.Parse()
	return opts
}

func (o options) validate() error {
	if o.steps < 0 {
		return fmt.Errorf("-steps must be >= 0")
	}

	switch o.direction {
	case "up", "down", "version":
		return nil
	case "force", "goto":
		if o.version < 0 {
			return fmt.Errorf("-version must be >= 0 for %q", o.direction)
		}
		return nil
	default:
		return fmt.Errorf("unknown direction %q (want up, down, version, force, goto)", o.direction)
	}
}

func newMigrator(cfg config.DatabaseConfig, migrationsPath string) (*migrate.Migrate, error) {
	dbURL := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:     "/" + cfg.DatabaseName,
		RawQuery: "sslmode=disable",
	}

	m, err := migrate.New("file://"+migrationsPath, dbURL.String())
	if err != nil {
		return nil, err
	}
	return m, nil
}

func applyMigration(m *migrate.Migrate, opts options) error {
	var err error

	switch opts.direction {
	case "up":
		if opts.steps > 0 {
			err = m.Steps(opts.steps)
		} else {
			err = m.Up()
		}

	case "down":
		if opts.steps > 0 {
			err = m.Steps(-opts.steps)
		} else {
			err = m.Down()
		}

	case "version":
		return printVersion(m)

	case "force":
		err = m.Force(opts.version)

	case "goto":
		err = m.Migrate(uint(opts.version))
	}

	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("no change")
		return nil
	}
	return err
}

func printVersion(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("version: none")
		fmt.Println("dirty: false")
		return nil
	}
	if err != nil {
		return fmt.Errorf("get migration version: %w", err)
	}
	fmt.Printf("version: %d\ndirty: %t\n", version, dirty)
	return nil
}