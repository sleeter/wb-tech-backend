package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/avast/retry-go/v4"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"time"
	"wb-tech-backend/internal/core"
	"wb-tech-backend/internal/http_server"
	"wb-tech-backend/internal/nats"
	"wb-tech-backend/internal/pkg/config"
	"wb-tech-backend/internal/repository"
	"wb-tech-backend/internal/service"
)

func main() {
	ctx := context.Background()
	loader := config.PrepareLoader(config.WithConfigPath("./config.yaml"))

	cfg, err := core.ParseConfig(loader)
	if err != nil {
		log.Fatalf("Failed to parse config: %s", err)
	}

	err = retry.Do(func() error {
		return UpMigrations(cfg)
	}, retry.Attempts(4), retry.Delay(2*time.Second))
	if err != nil {
		log.Fatalf(err.Error())
	}
	repo, err := repository.NewRepository(ctx, cfg)
	if err != nil {
		log.Fatalf("Init repository: %s", err)
	}

	serv := service.NewService(repo, cfg)
	n, err := nats.NewNats(serv)
	if err != nil {
		log.Fatalf("Init nats: %s", err)
	}
	go n.SubscribeToUpdates(ctx)

	app := http_server.New(serv)

	if err := app.Start(ctx); err != nil {
		log.Fatalf(err.Error())
	}
}
func UpMigrations(cfg *core.Config) error {
	db, err := sql.Open("pgx", cfg.Storage.URL)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
