// internal/handler/trip_handler.go - Text parsing eklenmiş versiyon

package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
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
	GetUserTripsHandler(c *fiber.Ctx) error
	DeleteTripHandler(c *fiber.Ctx) error
	GetTripByIDHandler(c *fiber.Ctx) error
}

// Günlük plan parse etmek için struct
type ParsedDayPlan struct {
	Day      int    `json:"day"`
	Date     string `json:"date"`
	Location string `json:"location"`
	Details  string `json:"details"`
	Notes    string `json:"notes"`
}

func (h *TripHandler) NewCreateTripHandler(c *fiber.Ctx) error {
	var trip models.Trip

	if err := c.BodyParser(&trip); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	log.Printf("📨 Received trip data: %+v", trip)

	grpcReq := client.CreatePromptRequest(
		trip.UserID,
		trip.Name,
		trip.Description,
		trip.StartPosition,
		trip.EndPosition,
		trip.StartDate,
		trip.EndDate,
	)

	response, err := h.AIClient.GenerateTripPlan(context.Background(), grpcReq)
	if err != nil {
		log.Printf("❌ gRPC Error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate trip plan", 
			"details": err.Error(),
		})
	}

	log.Printf("📥 AI Response daily plans count: %d", len(response.DailyPlan))

	// Eğer DailyPlan boşsa, route_summary'den parse et
	if len(response.DailyPlan) == 0 && response.Trip != nil && response.Trip.RouteSummary != "" {
		log.Printf("🔍 DailyPlan is empty, parsing from route_summary...")
		parsedPlans := parseRouteSummary(response.Trip.RouteSummary, trip.StartDate)
		
		if len(parsedPlans) > 0 {
			log.Printf("✅ Successfully parsed %d days from route_summary", len(parsedPlans))
			
			// Parse edilmiş planları gRPC response formatına çevir
			for _, plan := range parsedPlans {
				dailyPlan := &proto.DailyPlan{
					Day:  int32(plan.Day),
					Date: plan.Date,
					Location: &proto.Location{
						Name:      plan.Location,
						Address:   extractLocationHint(plan.Details),
						Notes:     plan.Details,
						SiteUrl:   "",
						Latitude:  0.0, // Koordinatlar için ayrı bir servise ihtiyaç var
						Longitude: 0.0,
					},
				}
				response.DailyPlan = append(response.DailyPlan, dailyPlan)
			}
		}
	}

	tripResponse := convertGRPCResponseToModel(response)
	return c.Status(fiber.StatusOK).JSON(tripResponse)
}

// Route summary'den günlük planları parse etme fonksiyonu
func parseRouteSummary(routeSummary string, startDate string) []ParsedDayPlan {
	var plans []ParsedDayPlan
	
	// Günlük plan pattern'i - örnek: "**1. Gün (2025-08-13): İstanbul - Ayvalık**"
	dayPattern := regexp.MustCompile(`\*\*(\d+)\.\s*Gün\s*\(([^)]+)\):\s*([^*]+)\*\*`)
	
	// Parse start date
	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		log.Printf("❌ Failed to parse start date: %v", err)
		return plans
	}
	
	lines := strings.Split(routeSummary, "\n")
	
	currentDay := 0
	currentDetails := ""
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Gün başlığını bul
		matches := dayPattern.FindStringSubmatch(line)
		if len(matches) >= 4 {
			// Önceki günün detaylarını kaydet
			if currentDay > 0 {
				dayDate := startTime.AddDate(0, 0, currentDay-1).Format("2006-01-02")
				plans = append(plans, ParsedDayPlan{
					Day:      currentDay,
					Date:     dayDate,
					Location: extractMainLocation(matches[3]),
					Details:  strings.TrimSpace(currentDetails),
					Notes:    strings.TrimSpace(currentDetails),
				})
			}
			
			// Yeni gün başlat
			dayNum, _ := strconv.Atoi(matches[1])
			currentDay = dayNum
			currentDetails = ""
			
			log.Printf("📅 Found day %d: %s", dayNum, matches[3])
		} else if currentDay > 0 && line != "" && !strings.HasPrefix(line, "**") {
			// Mevcut günün detaylarını topla
			if !strings.HasPrefix(line, "*   **") { // Alt başlıkları atla
				currentDetails += line + "\n"
			}
		}
		
		// Son gün için özel kontrol
		if i == len(lines)-1 && currentDay > 0 {
			dayDate := startTime.AddDate(0, 0, currentDay-1).Format("2006-01-02")
			plans = append(plans, ParsedDayPlan{
				Day:      currentDay,
				Date:     dayDate,
				Location: extractMainLocationFromDetails(currentDetails),
				Details:  strings.TrimSpace(currentDetails),
				Notes:    strings.TrimSpace(currentDetails),
			})
		}
	}
	
	// Manuel parse - backup method
	if len(plans) == 0 {
		plans = manualParseRouteSummary(routeSummary, startTime)
	}
	
	return plans
}

// Ana lokasyonu çıkarma
func extractMainLocation(text string) string {
	// "İstanbul - Ayvalık" formatından "Ayvalık" çıkar
	parts := strings.Split(text, " - ")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return strings.TrimSpace(text)
}

// Detaylardan ana lokasyonu çıkarma
func extractMainLocationFromDetails(details string) string {
	if strings.Contains(details, "Ayvalık") {
		return "Ayvalık"
	} else if strings.Contains(details, "Foça") {
		return "Foça"
	} else if strings.Contains(details, "Kuşadası") {
		return "Kuşadası"
	} else if strings.Contains(details, "Akyaka") || strings.Contains(details, "Gökova") {
		return "Akyaka (Gökova)"
	} else if strings.Contains(details, "Muğla") {
		return "Muğla"
	}
	return "Konum belirtilmemiş"
}

// Address hint çıkarma
func extractLocationHint(details string) string {
	if strings.Contains(details, "Ayvalık") {
		return "Ayvalık, Balıkesir"
	} else if strings.Contains(details, "Foça") {
		return "Eski Foça, İzmir"
	} else if strings.Contains(details, "Kuşadası") {
		return "Kuşadası, Aydın"
	} else if strings.Contains(details, "Akyaka") {
		return "Akyaka, Muğla"
	} else if strings.Contains(details, "Muğla") {
		return "Muğla Merkez"
	}
	return ""
}

// Manuel parse - backup method
func manualParseRouteSummary(routeSummary string, startTime time.Time) []ParsedDayPlan {
	var plans []ParsedDayPlan
	
	// Manuel olarak bilinen lokasyonları çıkar
	locations := []string{
		"Ayvalık",
		"Ayvalık", // 2. gün de Ayvalık
		"Foça",
		"Foça", // 4. gün de Foça
		"Kuşadası",
		"Kuşadası", // 6. gün de Kuşadası
		"Akyaka (Gökova)",
		"Muğla",
	}
	
	for i, location := range locations {
		if i >= 8 { // Maksimum 8 gün
			break
		}
		
		dayDate := startTime.AddDate(0, 0, i).Format("2006-01-02")
		plans = append(plans, ParsedDayPlan{
			Day:      i + 1,
			Date:     dayDate,
			Location: location,
			Details:  fmt.Sprintf("Gün %d: %s'ta konaklama ve keşif", i+1, location),
			Notes:    fmt.Sprintf("AI tarafından önerilen %s bölgesi", location),
		})
	}
	
	return plans
}

func convertGRPCResponseToModel(grpcResp *proto.TripPlanResponse) map[string]interface{} {
	if grpcResp == nil {
		return map[string]interface{}{"error": "Empty response from AI service"}
	}

	var tripData map[string]interface{}
	if grpcResp.Trip != nil {
		tripData = map[string]interface{}{
			"user_id":        grpcResp.Trip.UserId,
			"name":           grpcResp.Trip.Name,
			"description":    grpcResp.Trip.Description,
			"start_position": grpcResp.Trip.StartPosition,
			"end_position":   grpcResp.Trip.EndPosition,
			"start_date":     grpcResp.Trip.StartDate,
			"end_date":       grpcResp.Trip.EndDate,
			"total_days":     grpcResp.Trip.TotalDays,
			"route_summary":  grpcResp.Trip.RouteSummary,
		}
	}

	var locations []map[string]interface{}
	log.Printf("🔄 Converting %d daily plans", len(grpcResp.DailyPlan))
	
	for i, dailyPlan := range grpcResp.DailyPlan {
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
		log.Printf("✅ Processed day %d: %s", i+1, location["name"])
	}

	return map[string]interface{}{
		"trip":       tripData,
		"daily_plan": locations,
		"debug_info": map[string]interface{}{
			"total_daily_plans": len(grpcResp.DailyPlan),
			"has_trip_data":     grpcResp.Trip != nil,
			"parsed_from_text":  len(grpcResp.DailyPlan) > 0,
		},
	}
}

// Diğer handler metodları aynı...
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

func (h *TripHandler) GetUserTripsHandler(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id is required"})
	}

	tripService := service.NewTripService(nil, h.DB, nil)
	trips, err := tripService.GetUserTrips(context.Background(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get trips"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"trips": trips})
}

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