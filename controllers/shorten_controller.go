package controllers

import (
	"fmt"
	"net/http"
	"time"
	modelrequests "url-shortening-service/models/requests"
	"url-shortening-service/services"

	"github.com/labstack/echo/v4"
)

type ShortenController interface {
	Create(c echo.Context) error
	GetOriginalUrl(c echo.Context) error
	UpdateUrl(c echo.Context) error
	Delete(c echo.Context) error
	GetUrlStatistics(c echo.Context) error
	ShortenResult(c echo.Context) error
	ShortenResult2(c echo.Context) error
}

type ShortenControllerImplementation struct {
	ShortenService services.ShortenService
}

func NewShortenController(shortenService services.ShortenService) ShortenController {
	return &ShortenControllerImplementation{
		ShortenService: shortenService,
	}
}

func (controller *ShortenControllerImplementation) Create(c echo.Context) error {
	var createShortenRequest modelrequests.CreateShortenRequest
	err := c.Bind(&createShortenRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	now := time.Now().UnixMilli()
	httpCode, response := controller.ShortenService.Create(c.Request().Context(), createShortenRequest, now)
	return c.JSON(httpCode, response)
}

func (controller *ShortenControllerImplementation) GetOriginalUrl(c echo.Context) error {
	shortCodeParam := c.Param("shortCode")
	_, url, _ := controller.ShortenService.GetOriginalUrl(c.Request().Context(), shortCodeParam)
	fmt.Println("url:", url)
	return c.Redirect(http.StatusFound, url)
}

func (controller *ShortenControllerImplementation) UpdateUrl(c echo.Context) error {
	var updateUrlShortenRequest modelrequests.UpdateUrlShortenRequest
	err := c.Bind(&updateUrlShortenRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}
	shortCodeParam := c.Param("shortCode")
	now := time.Now().UnixMilli()
	httpCode, response := controller.ShortenService.UpdateUrl(c.Request().Context(), updateUrlShortenRequest, shortCodeParam, now)
	return c.JSON(httpCode, response)
}

func (controller *ShortenControllerImplementation) Delete(c echo.Context) error {
	shortCodeParam := c.Param("shortCode")
	httpCode, response := controller.ShortenService.Delete(c.Request().Context(), shortCodeParam)
	return c.JSON(httpCode, response)
}

func (controller *ShortenControllerImplementation) GetUrlStatistics(c echo.Context) error {
	shortCodeParam := c.Param("shortCode")
	httpCode, response := controller.ShortenService.GetUrlStatistics(c.Request().Context(), shortCodeParam)
	return c.JSON(httpCode, response)
}

func (controller *ShortenControllerImplementation) ShortenResult(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "redirect result",
	})
}

func (controller *ShortenControllerImplementation) ShortenResult2(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "shorten result 2",
	})
}
