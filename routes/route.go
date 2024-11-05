package routes

import (
	"url-shortening-service/controllers"

	"github.com/labstack/echo/v4"
)

func ShortenRoute(e *echo.Echo, controller controllers.ShortenController) {
	e.POST("/shorten", controller.Create)
	e.GET("/shorten/:shortCode", controller.GetOriginalUrl)
	e.PUT("/shorten/:shortCode", controller.UpdateUrl)
	e.DELETE("/shorten/:shortCode", controller.Delete)
	e.GET("/shorten/:shortCode/stats", controller.GetUrlStatistics)
	e.GET("/shorten-result", controller.ShortenResult)
	e.GET("/shorten-result2", controller.ShortenResult2)
}
