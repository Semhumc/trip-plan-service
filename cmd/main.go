// main.go

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

func main() {
	// Değişkenler artık .env dosyasından env_file ile container'a doğru şekilde aktarılacak
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_DATABASE")
	host := os.Getenv("DB_HOST") // Değeri: "trip-location-service-psql"
	port := os.Getenv("DB_PORT") // Değeri: "5432"
	appPort := os.Getenv("PORT") // Değeri: "8085"
	aiServiceAddr := os.Getenv("AI_SERVICE_ADDR")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Bağlantı dizesi artık doğru değerlerle oluşturulacak
	dburl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)

	db, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Fatalf("Veritabanı bağlantı hatası: %v", err)
	}
	defer db.Close()

	// Bağlantıyı doğrulamak için ping atın
	if err = db.Ping(); err != nil {
		log.Fatalf("Veritabanına ping atılamadı: %v", err)
	}
	log.Println("Veritabanı bağlantısı başarıyla sağlandı!")

	aiClient, err := client.NewAIClient(aiServiceAddr)
	if err != nil {
		log.Fatalf("AI istemcisi oluşturulamadı: %v", err)
	}
	defer aiClient.Close()

	tripHandler := handler.NewTripHandler(db, aiClient)
	routes.TripRoutes(app, tripHandler)

	if appPort == "" {
		appPort = "8085" // Ortam değişkeni yoksa varsayılan port
	}

	log.Printf("Sunucu :%s portunda dinleniyor...", appPort)
	if err := app.Listen(":" + appPort); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
