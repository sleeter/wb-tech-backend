package http_server

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"wb-tech-backend/internal/http_server/handlers"
	"wb-tech-backend/internal/pkg/web"
	"wb-tech-backend/internal/service"
)

type App struct {
	Server  web.Server
	Router  *gin.Engine
	Service *service.Service
}

func New(service *service.Service) *App {
	app := &App{
		Service: service,
	}
	app.initRoutes()
	app.Server = web.NewServer(service.Config.Server, app.Router)
	return app
}

func (app *App) Start(ctx context.Context) error {
	return app.Server.Run(ctx)
}

func (app *App) initRoutes() {
	app.Router = gin.Default()

	app.Router.GET("/order", app.mappedHandler(handlers.GetOrder))
	app.Router.GET("/orders", app.mappedHandler(handlers.GetOrders))
}

func (app *App) mappedHandler(handler func(*gin.Context, *service.Service) error) gin.HandlerFunc {

	return func(ctx *gin.Context) {

		if err := handler(ctx, app.Service); err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		}
	}
}