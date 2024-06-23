package handlers

import (
	"log/slog"
	"net/http"

	"wb-tech-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func GetOrder(ctx *gin.Context, service *service.Service) error {
	var param struct {
		OrderId string `json:"order_uid" binding:"required"`
	}

	if err := ctx.BindJSON(&param); err != nil {
		slog.Debug("Error with getting order: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return nil
	}
	order, err := service.GetOrder(ctx, param.OrderId)
	if err != nil {
		slog.Debug("Error with getting order: %s", err)
		return err
	}
	ctx.JSON(http.StatusOK, order)
	return nil
}
func GetOrder2(ctx *gin.Context, service *service.Service) error {
	orderId := ctx.Query("order_uid")
	if orderId == "" {
		slog.Debug("Error with getting order: order id is empty")
	}
	order, err := service.GetOrder(ctx, orderId)
	if err != nil {
		slog.Debug("Error with getting order: %s", err)
		return err
	}
	ctx.JSON(http.StatusOK, order)
	return nil
}
func GetOrders(ctx *gin.Context, service *service.Service) error {
	orders, err := service.ListOfOrders(ctx)
	if err != nil {
		slog.Debug("Error with getting orders: %s", err)
		return err
	}
	ctx.JSON(http.StatusOK, orders)
	return nil
}
