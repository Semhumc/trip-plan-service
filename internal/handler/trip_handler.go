package handler

import (
	"context"
	"database/sql"
	"log"
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
		log.Printf("❌ Body parse hatası: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	log.Printf("📨 Received trip data: %+v", trip)

	// gRPC request oluştur
	grpcReq := client.CreatePromptRequest(
		trip.UserID,
		trip.Name,
		trip.Description,
		trip.StartPosition,
		trip.EndPosition,
		trip.StartDate,
		trip.EndDate,
	)

	log.Printf("📤 gRPC request gönderiliyor: %+v", grpcReq)

	// AI servisini çağır
	response, err := h.AIClient.GenerateTripPlan(context.Background(), grpcReq)
	if err != nil {
		log.Printf("❌ gRPC Error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to generate trip plan",
			"details": err.Error(),
		})
	}

	log.Printf("📥 gRPC Response alındı - Daily plans count: %d", len(response.TripOptions))

	// gRPC response'u frontend için uygun formata çevir
	tripResponse := convertTripOptionsToModel(response)

	log.Printf("✅ Response hazırlandı: %+v", tripResponse)
	return c.Status(fiber.StatusOK).JSON(tripResponse)
}

// gRPC response'u frontend modelına çevir
func convertTripOptionsToModel(grpcResp *proto.TripOptionsResponse) map[string]interface{} {
	if grpcResp == nil {
		return map[string]interface{}{"error": "Empty response from AI service"}
	}

	var tripOptions []map[string]interface{}

	for _, option := range grpcResp.TripOptions {
		// Trip data
		var tripData map[string]interface{}
		if option.Trip != nil {
			tripData = map[string]interface{}{
				"user_id":        option.Trip.UserId,
				"name":           option.Trip.Name,
				"description":    option.Trip.Description,
				"start_position": option.Trip.StartPosition,
				"end_position":   option.Trip.EndPosition,
				"start_date":     option.Trip.StartDate,
				"end_date":       option.Trip.EndDate,
				"total_days":     option.Trip.TotalDays,
			}
		}

		// Daily plans
		var locations []map[string]interface{}
		for _, dailyPlan := range option.DailyPlan {
			location := map[string]interface{}{
				"day":  dailyPlan.Day,
				"date": dailyPlan.Date,
			}

			if dailyPlan.Location != nil {
				location["name"] = dailyPlan.Location.Name
				location["address"] = dailyPlan.Location.Address
				location["site_url"] = dailyPlan.Location.SiteUrl
				location["latitude"] = dailyPlan.Location.Latitude
				location["longitude"] = dailyPlan.Location.Longitude
				location["notes"] = dailyPlan.Location.Notes
			}

			locations = append(locations, location)
		}

		// Option oluştur
		tripOption := map[string]interface{}{
			"theme":       option.Theme,
			"description": option.Description,
			"trip":        tripData,
			"daily_plan":  locations,
		}

		tripOptions = append(tripOptions, tripOption)
	}

	return map[string]interface{}{
		"trip_options":  tripOptions,
		"total_options": len(tripOptions),
	}
}

func (h *TripHandler) SaveTripHandler(c *fiber.Ctx) error {
	var trip models.TripWithLocations

	if err := c.BodyParser(&trip); err != nil {
		log.Printf("❌ Save trip body parse hatası: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	log.Printf("💾 Saving trip: %s with %d locations", trip.Trip.Name, len(trip.Locations))

	tripService := service.NewTripService(&trip.Trip, h.DB, trip.Locations)

	err := tripService.SaveTripWLocations(context.Background())
	if err != nil {
		log.Printf("❌ Trip save hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save trip"})
	}

	log.Printf("✅ Trip başarıyla kaydedildi")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "trip saved successfully"})
}

func (h *TripHandler) GetUserTripsHandler(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id is required"})
	}

	log.Printf("📖 Getting trips for user: %s", userID)

	tripService := service.NewTripService(nil, h.DB, nil)
	trips, err := tripService.GetUserTrips(context.Background(), userID)
	if err != nil {
		log.Printf("❌ Get user trips hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get trips"})
	}

	log.Printf("✅ %d trip bulundu", len(trips))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"trips": trips})
}

func (h *TripHandler) DeleteTripHandler(c *fiber.Ctx) error {
	tripIDStr := c.Params("id")
	tripID, err := strconv.Atoi(tripIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid trip id"})
	}

	log.Printf("🗑️ Deleting trip: %d", tripID)

	tripService := service.NewTripService(nil, h.DB, nil)
	err = tripService.DeleteTrip(context.Background(), int32(tripID))
	if err != nil {
		log.Printf("❌ Delete trip hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete trip"})
	}

	log.Printf("✅ Trip silindi: %d", tripID)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "trip deleted successfully"})
}

func (h *TripHandler) GetTripByIDHandler(c *fiber.Ctx) error {
	tripIDStr := c.Params("id")
	tripID, err := strconv.Atoi(tripIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid trip id"})
	}

	log.Printf("📖 Getting trip by ID: %d", tripID)

	tripService := service.NewTripService(nil, h.DB, nil)
	trip, err := tripService.GetTripByID(context.Background(), int32(tripID))
	if err != nil {
		log.Printf("❌ Get trip by ID hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get trip"})
	}

	log.Printf("✅ Trip bulundu: %s", trip.Trip.Name)
	return c.Status(fiber.StatusOK).JSON(trip)
}
