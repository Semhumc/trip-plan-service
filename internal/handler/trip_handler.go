package handler

import (
	"context"
	"database/sql"
	"time"

	"trip-plan-service/internal/client"
	"trip-plan-service/internal/models"
	"trip-plan-service/internal/service"

	"github.com/Semhumc/grpc-proto/proto"
	"github.com/gofiber/fiber/v2"
)

type TripHandler struct {
	DB       *sql.DB
	AIClient *client.AIClient
}

func NewTripHandler(db *sql.DB, aiClient *client.AIClient) *TripHandler {
	return &TripHandler{
		DB:       db,
		AIClient: aiClient,
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

	// Create gRPC request
	grpcReq := client.CreatePromptRequest(
		trip.UserID,
		trip.Name,
		trip.Description,
		trip.StartPosition,
		trip.EndPosition,
		trip.StartDate.Format(time.RFC3339),
		trip.EndDate.Format(time.RFC3339),
	)

	// Call AI service via gRPC
	response, err := h.AIClient.GenerateTripPlan(context.Background(), grpcReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate trip plan", "details": err.Error()})
	}

	// Convert gRPC response to your internal model
	tripResponse := convertGRPCResponseToModel(response)

	return c.Status(fiber.StatusOK).JSON(tripResponse)
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

func convertGRPCResponseToModel(grpcResp *proto.TripPlanResponse) map[string]interface{} {
	// Convert locations
	locations := make([]map[string]interface{}, len(grpcResp.DailyPlan))
	for i, dailyPlan := range grpcResp.DailyPlan {
		location := map[string]interface{}{
			"day":       dailyPlan.Day,
			"date":      dailyPlan.Date,
			"name":      dailyPlan.Location.Name,
			"address":   dailyPlan.Location.Address,
			"site_url":  dailyPlan.Location.SiteUrl,
			"latitude":  dailyPlan.Location.Latitude,
			"longitude": dailyPlan.Location.Longitude,
			"notes":     dailyPlan.Location.Notes,
		}
		locations[i] = location
	}

	return map[string]interface{}{
		"trip": map[string]interface{}{
			"user_id":        grpcResp.Trip.UserId,
			"name":           grpcResp.Trip.Name,
			"description":    grpcResp.Trip.Description,
			"start_position": grpcResp.Trip.StartPosition,
			"end_position":   grpcResp.Trip.EndPosition,
			"start_date":     grpcResp.Trip.StartDate,
			"end_date":       grpcResp.Trip.EndDate,
			"total_days":     grpcResp.Trip.TotalDays,
			"route_summary":  grpcResp.Trip.RouteSummary,
		},
		"daily_plan": locations,
	}
}
