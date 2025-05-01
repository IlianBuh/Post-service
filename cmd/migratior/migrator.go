package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"

	"github.com/IlianBuh/Post-service/internal/config"
	"github.com/IlianBuh/Post-service/internal/lib/logger/sl"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		migrationsPath string
	)

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to directory with migration files")
	cfg := config.New().Storage

	conn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	m, err := migrate.New(
		"file://"+migrationsPath,
		conn,
	)
	if err != nil {
		slog.Error("failed to create new migrator instance", sl.Err(err))
		return
	}
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no changes")
			return
		}
		slog.Error("failed to migrate", sl.Err(err))
		return
	}
}
