package handler

import (
	"context"
	"database/sql"
	"strconv"
	

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
	GetUserTripsHandler(c *fiber.Ctx) error
	DeleteTripHandler(c *fiber.Ctx) error
	GetTripByIDHandler(c *fiber.Ctx) error
	
}


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
		trip.StartDate,
		trip.EndDate,
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


// YENİ: Kullanıcının tüm triplerini getir
func (h *TripHandler) GetUserTripsHandler(c *fiber.Ctx) error {
	// Query parametresinden veya header'dan user_id'yi al
	userID := c.Query("user_id")
	if userID == "" {
		// Alternatif olarak session'dan veya JWT'den user_id alınabilir
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id is required"})
	}

	tripService := service.NewTripService(nil, h.DB, nil)
	trips, err := tripService.GetUserTrips(context.Background(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get trips"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"trips": trips})
}

// YENİ: Trip sil
func (h *TripHandler) DeleteTripHandler(c *fiber.Ctx) error {
	tripIDStr := c.Params("id")
	tripID, err := strconv.Atoi(tripIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid trip id"})
	}

	tripService := service.NewTripService(nil, h.DB, nil)
	err = tripService.DeleteTrip(context.Background(), int32(tripID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete trip"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "trip deleted successfully"})
}

// YENİ: ID'ye göre trip getir
func (h *TripHandler) GetTripByIDHandler(c *fiber.Ctx) error {
	tripIDStr := c.Params("id")
	tripID, err := strconv.Atoi(tripIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid trip id"})
	}

	tripService := service.NewTripService(nil, h.DB, nil)
	trip, err := tripService.GetTripByID(context.Background(), int32(tripID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get trip"})
	}

	return c.Status(fiber.StatusOK).JSON(trip)
}
