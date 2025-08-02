// internal/routes/trip_route.go - Güncellenmiş rotalar

package routes

import (
	"trip-plan-service/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func TripRoutes(router fiber.Router, handler handler.TripHandlerInterface) {
	api := router.Group("/api/v1/trip")

	// Mevcut endpoint'ler
	api.Post("/preview", handler.NewCreateTripHandler) 
	api.Post("/save", handler.SaveTripHandler)
	
	// YENİ endpoint'ler
	api.Get("/list", handler.GetUserTripsHandler)        // Kullanıcı triplerini listele
	api.Get("/:id", handler.GetTripByIDHandler)          // ID'ye göre trip getir
	api.Delete("/:id", handler.DeleteTripHandler)        // Trip sil
}