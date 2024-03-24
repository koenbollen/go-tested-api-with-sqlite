package internal

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"

	"github.com/koenbollen/go-tested-api-with-sqlite/migrations"
)

type Config struct {
	DSN string `default:":memory:?cache=shared"`
}

type Dependencies struct {
	DB *sql.DB
}

func DefaultConfig() *Config {
	cfg := &Config{}
	t := reflect.ValueOf(cfg).Elem()
	for i := 0; i < t.Type().NumField(); i++ {
		f := t.Type().Field(i)
		if def, ok := f.Tag.Lookup("default"); ok {
			t.Field(i).SetString(def)
		}
	}
	return cfg
}

func ConfigFromEnv(ctx context.Context) (*Config, error) {
	cfg := DefaultConfig()
	if v, ok := os.LookupEnv("DSN"); ok {
		cfg.DSN = v
	}
	return cfg, nil
}

func Setup(ctx context.Context, config *Config) (*Dependencies, error) {
	var err error
	deps := &Dependencies{}

	if deps.DB, err = sql.Open("sqlite", config.DSN); err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := migrations.Up(ctx, deps.DB); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	go func() {
		<-ctx.Done()
		deps.DB.Close()
	}()

	return deps, nil
}
