package routes

import (
	"trip-plan-service/internal/handler"

	"github.com/gofiber/fiber/v2"
)


func TripRoutes(router fiber.Router, handler handler.TripHandlerInterface) {
	api := router.Group("/api/v1/trip")

	api.Post("/preview", handler.NewCreateTripHandler) 
	api.Post("/save", handler.SaveTripHandler)         
}