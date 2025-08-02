package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"trip-plan-service/internal/client"
	"trip-plan-service/internal/handler"
	"trip-plan-service/internal/routes"

	_ "github.com/lib/pq"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var (
	user            = os.Getenv("DB_USERNAME")
	password        = os.Getenv("DB_PASSWORD")
	dbname          = os.Getenv("DB_DATABASE")
	host            = os.Getenv("DB_HOST")
	port            = os.Getenv("DB_PORT")
	//aiServiceAddr   = os.Getenv("AI_SERVICE_ADDR") // e.g., "localhost:50051"
	tripServicePort = os.Getenv("TRIP_SERVICE_PORT")

	aiServiceAddr = "localhost:50051"
)

func main() {
	app := fiber.New()

	// ✅ CORS Middleware
	app.Use(cors.New(cors.Config{
        AllowOrigins: "http://localhost:3000", // frontend adresin
        AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
        AllowHeaders: "Origin, Content-Type, Accept, Authorization",
        AllowCredentials: true,
    }))

	dburl := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		user, password, dbname, host, port)

	db, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Fatalf("Veritabanı bağlantı hatası: %v", err)
	}
	defer db.Close()

	aiClient, err := client.NewAIClient(aiServiceAddr)
	if err != nil {
		log.Fatalf("Failed to create AI client: %v", err)
	}
	defer aiClient.Close()

	tripHandler := handler.NewTripHandler(db, aiClient)

	routes.TripRoutes(app, tripHandler)

	// ✅ Dinlenecek port .env'den alınabilir
	if tripServicePort == "" {
		tripServicePort = "8000"
	}
	err = app.Listen(":" + tripServicePort)
	if err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
