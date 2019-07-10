package migrations

//go:generate ./gen.sh

import (
	"github.com/gobuffalo/packr"
	migrate "github.com/rubenv/sql-migrate"
)

var (
	// Migrations to setup the database
	Migrations = &migrate.PackrMigrationSource{
		Box: packr.NewBox("."),
	}
)
