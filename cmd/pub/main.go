package main

import (
	"github.com/nats-io/stan.go"
	"log"
	"os"
	"wb-tech-backend/internal/core"
	"wb-tech-backend/internal/pkg/config"
)

func main() {
	loader := config.PrepareLoader(config.WithConfigPath("config.yml"))
	cfg, err := core.ParseConfig(loader)
	if err != nil {
		log.Fatalf("Error with read config: %s", err)
	}
	sc, err := stan.Connect("test-cluster", cfg.Nats.Prod, stan.NatsURL(cfg.Nats.PubUrl))
	if err != nil {
		log.Fatalf("Error with nats connection: %s", err)
	}
	defer func(sc stan.Conn) {
		err := sc.Close()
		if err != nil {
			log.Fatalf("Error with close nats connection: %s", err)
		}
	}(sc)

	bytes, err := os.ReadFile("json_models/model.json")

	err = sc.Publish(cfg.Nats.Subject, bytes)
	if err != nil {
		log.Fatalf("Error with publish to nats: %s", err)
	}
	log.Printf("first message successfuly publish to nats")

	bytes, err = os.ReadFile("json_models/model2.json")

	err = sc.Publish(cfg.Nats.Subject, bytes)
	if err != nil {
		log.Fatalf("Error with publish to nats: %s", err)
	}
	log.Printf("second message successfuly publish to nats")

	bytes, err = os.ReadFile("json_models/model3.json")

	err = sc.Publish(cfg.Nats.Subject, bytes)
	if err != nil {
		log.Fatalf("Error with publish to nats: %s", err)
	}
	log.Printf("third message successfuly publish to nats")
}
