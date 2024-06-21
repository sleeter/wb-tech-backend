package service

import (
	"context"
	"fmt"
	"wb-tech-backend/internal/core"
	"wb-tech-backend/internal/models"
	"wb-tech-backend/internal/repository"
)

type Repository interface {
	AddOrder(ctx context.Context, order models.Order) error
	GetOrderById(ctx context.Context, orderId string) (models.Order, error)
	GetOrders(ctx context.Context) ([]models.Order, error)
}

type Deps struct {
	Repository *repository.Repository
	Config     *core.Config
}

type Service struct {
	Deps
}

func NewService(r *repository.Repository, cfg *core.Config) *Service {
	return &Service{
		Deps{
			Repository: r,
			Config:     cfg,
		}}
}

func (s Service) AddOrder(ctx context.Context, order models.Order) error {
	return s.Repository.AddOrder(ctx, order)
}
func (s Service) ListOfOrders(ctx context.Context) ([]models.Order, error) {
	return s.Repository.GetOrders(ctx)
}
func (s Service) GetOrder(ctx context.Context, orderId string) (models.Order, error) {
	order, ok := s.Repository.Cash[orderId]
	if !ok {
		return models.Order{}, fmt.Errorf("order with id=%d not exsits", orderId)
	}
	return order, nil
}
