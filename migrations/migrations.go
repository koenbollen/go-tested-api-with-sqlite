package migrations

import (
	"context"
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrations embed.FS

func Up(ctx context.Context, db *sql.DB) error {
	source, err := iofs.New(migrations, ".")
	if err != nil {
		return err
	}

	database, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", source, "sqlite", database)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
