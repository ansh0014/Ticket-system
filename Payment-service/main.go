package main

import (
    "log"
    "net/http"
    "os"

    "github.com/ansh0014/payment/config"
    "github.com/ansh0014/payment/router"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    godotenv.Load("D:\\Ticket-System\\Ticket-system\\Payment-service\\.env")

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
        port = "8003" // Default port for Payment Service
    }

    log.Printf("Payment service starting on port %s...", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}