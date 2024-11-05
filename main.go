package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"
	"url-shortening-service/controllers"
	"url-shortening-service/repositories"
	"url-shortening-service/routes"
	"url-shortening-service/services"
	"url-shortening-service/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func main() {
	postgresUtil := utils.NewPostgresConnection()
	e := echo.New()
	validate := validator.New()

	shortenRepository := repositories.NewShortenRepository()
	shortenService := services.NewShortenService(postgresUtil, validate, shortenRepository)
	shortenController := controllers.NewShortenController(shortenService)
	routes.ShortenRoute(e, shortenController)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
