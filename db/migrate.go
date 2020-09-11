package db

import (
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"os"
)

const (
	migrateConnectionString = `postgres://%s:%s/%s?user=%s&password=%s&sslmode=%s&sslcert=%s&sslkey=%s&sslrootcert=%s&connect_timeout=60`
)

func Migrate() {
	logbuch.Info("Migrating database schema (if required)")
	m, err := migrate.New("file://schema", getMigrationConnectionString())

	if err != nil {
		logbuch.Fatal("Error migrating database schema", logbuch.Fields{"err": err})
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logbuch.Fatal("Error migrating database schema", logbuch.Fields{"err": err})
		return
	}

	if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
		logbuch.Fatal("Error migrating database schema", logbuch.Fields{"source_err": sourceErr, "db_err": dbErr})
	}

	logbuch.Info("Done migrating database schema")
}

func getMigrationConnectionString() string {
	host := os.Getenv("MB_DB_HOST")
	port := os.Getenv("MB_DB_PORT")
	user := os.Getenv("MB_DB_USER")
	password := os.Getenv("MB_DB_PASSWORD")
	schema := os.Getenv("MB_DB_SCHEMA")
	sslMode := os.Getenv("MB_DB_SSLMODE")
	sslCert := os.Getenv("MB_DB_SSLCERT")
	sslKey := os.Getenv("MB_DB_SSLKEY")
	sslRootCert := os.Getenv("MB_DB_SSLROOTCERT")
	return fmt.Sprintf(migrateConnectionString, host, port, schema, user, password, sslMode, sslCert, sslKey, sslRootCert)
}
