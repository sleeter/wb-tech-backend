package nats

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"wb-tech-backend/internal/models"
	"wb-tech-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/stan.go"
)

type Deps struct {
	NatsConnection stan.Conn
	Service        *service.Service
}

type Nats struct {
	Deps
}

func NewNats(service *service.Service, clusterId, clientId, natsUrl string) (*Nats, error) {
	sc, err := stan.Connect(
		clusterId,
		clientId,
		stan.NatsURL(natsUrl))
	if err != nil {
		return nil, err
	}
	return &Nats{
		Deps{
			NatsConnection: sc,
			Service:        service,
		}}, nil
}

func (n *Nats) SubscribeToUpdates(wg *sync.WaitGroup, ctx context.Context, subj string) error {
	defer wg.Done()
	sub, err := n.NatsConnection.Subscribe(subj, func(msg *stan.Msg) {
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
		if err := validate.Struct(order.Delivery); err != nil {
			log.Printf("Validation error: %v", err)
			return
		}
		if err := validate.Struct(order.Payment); err != nil {
			log.Printf("Validation error: %v", err)
			return
		}
		for _, item := range order.Items {
			if err := validate.Struct(item); err != nil {
				log.Printf("Validation error: %v", err)
				return
			}
		}
		err := n.Service.AddOrder(ctx, order)
		if err != nil {
			log.Printf(err.Error())
		}
	})
	if err != nil {
		return err
	}
	for {
		if !sub.IsValid() {
			wg.Done()
			break
		}
	}
	err = sub.Unsubscribe()
	if err != nil {
		return err
	}
	return nil
}
