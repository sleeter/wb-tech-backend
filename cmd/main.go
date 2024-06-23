package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"sync"
	"time"

	"wb-tech-backend/internal/core"
	"wb-tech-backend/internal/http_server"
	"wb-tech-backend/internal/nats"
	"wb-tech-backend/internal/pkg/config"
	"wb-tech-backend/internal/repository"
	"wb-tech-backend/internal/service"

	"github.com/avast/retry-go/v4"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	ctx := context.Background()
	loader := config.PrepareLoader(config.WithConfigPath("./config.yml"))

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

	n, err := nats.NewNats(serv, "test-cluster", cfg.Nats.Sub, cfg.Nats.SubUrl)
	if err != nil {
		log.Fatalf("Init nats: %s", err)
	}
	defer func() {
		err := n.NatsConnection.Close()
		slog.Debug("Error with close nats: %s", err)
		return
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := n.SubscribeToUpdates(&wg, ctx, cfg.Nats.Subject)
		if err != nil {
			slog.Debug("Error with nats: %s", err)
		}
	}()

	app := http_server.New(serv)

	if err := app.Start(ctx); err != nil {
		log.Fatalf(err.Error())
	}
	wg.Wait()
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
