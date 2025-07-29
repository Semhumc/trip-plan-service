package handler

import (
	"context"
	"database/sql"

	"trip-plan-service/internal/models"
	"trip-plan-service/internal/service"

	"github.com/gofiber/fiber/v2"
)

type TripHandler struct {
	DB *sql.DB
}

func NewTripHandler(db *sql.DB) *TripHandler {
	return &TripHandler{
		DB: db,
	}
}

type TripHandlerInterface interface {
	NewCreateTripHandler(c *fiber.Ctx) error
	SaveTripHandler(c *fiber.Ctx) error
}

// aı dan alıyorum
func (h *TripHandler) NewCreateTripHandler(c *fiber.Ctx) error {
	var trip models.Trip

	if err := c.BodyParser(&trip); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// AI mantıksal olarak burada çağrılabilir — öneri dönülecek

	return c.Status(fiber.StatusOK).JSON(trip)
}

// kayıt yapıyorum
func (h *TripHandler) SaveTripHandler(c *fiber.Ctx) error {
	var trip models.TripWithLocations

	if err := c.BodyParser(&trip); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	tripService := service.NewTripService(&trip.Trip, h.DB, trip.Locations)

	err := tripService.SaveTripWLocations(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save trip"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "trip saved successfully"})
}
