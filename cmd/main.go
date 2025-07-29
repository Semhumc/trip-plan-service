package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"trip-plan-service/internal/handler"
	"trip-plan-service/internal/routes"

	"github.com/gofiber/fiber/v2"
)

var(
	user = os.Getenv("DB_USERNAME")
    password = os.Getenv("DB_PASSWORD")
    dbname = os.Getenv("DB_DATABASE")
    host = os.Getenv("DB_HOST")
    port = os.Getenv("DB_PORT")
)

func main() {
	app := fiber.New()

	dburl := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", 
	user, password, dbname, host, port)

	db, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Fatalf("Veritabanı bağlantı hatası: %v", err)
	}
	defer db.Close()

	tripHandler := handler.NewTripHandler(db)

	routes.TripRoutes(app, tripHandler)

	err = app.Listen(":6000")
	if err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
