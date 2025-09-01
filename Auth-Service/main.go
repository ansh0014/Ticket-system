package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ansh0014/auth/config"
	"github.com/ansh0014/auth/router"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load("D:\\Ticket-System\\Ticket-system\\Auth-Service\\.env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
		// Continue execution, don't exit
	}

	config.InitRedis()
	if err := config.InitMongo(); err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}

	r := router.SetupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}
	log.Printf("Auth service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
