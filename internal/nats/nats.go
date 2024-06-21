package nats

import (
	"context"
	"encoding/json"
	"log"

	"wb-tech-backend/internal/models"
	"wb-tech-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
)

type Deps struct {
	NatsConnection *nats.Conn
	Service        *service.Service
}

type Nats struct {
	Deps
}

func NewNats(service *service.Service) (*Nats, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}
	return &Nats{
		Deps{
			NatsConnection: nc,
			Service:        service,
		}}, err
}

func (n *Nats) SubscribeToUpdates(ctx context.Context) {
	sub, err := n.NatsConnection.Subscribe("L0", func(msg *nats.Msg) {
		var order models.Order
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}
		validate := validator.New()
		if err := validate.Struct(order); err != nil {
			log.Printf("Validation error: %v", err)
			return
		}
		err := n.Service.AddOrder(ctx, order)
		if err != nil {
			log.Printf(err.Error())
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := sub.Drain(); err != nil {
		log.Fatal(err)
	}
}
