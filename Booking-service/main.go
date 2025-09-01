package main

import (
    "log"
    "net/http"
    "os"

    "github.com/ansh0014/booking/config"
    "github.com/ansh0014/booking/router"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    godotenv.Load("D:\\Ticket-System\\Ticket-system\\Booking-service\\.env")

    // Initialize Redis
    config.InitRedis()
    log.Println("Redis initialized")
    
    // Initialize MongoDB
    if err := config.InitMongo(); err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    log.Println("MongoDB initialized")
    
    // Set up router
    r := router.SetupRoutes()
    
    // Get port from environment or use default
    port := os.Getenv("PORT")
    if port == "" {
        port = "8082" // Default port for Booking Service
    }
    
    log.Printf("Booking service starting on port %s...", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}