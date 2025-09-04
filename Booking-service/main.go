package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ansh0014/booking/Platform/event"
	"github.com/ansh0014/booking/Platform/flight"
	"github.com/ansh0014/booking/Platform/movie"
	"github.com/ansh0014/booking/Platform/railway"
	"github.com/ansh0014/booking/config"
	"github.com/ansh0014/booking/router"
	"github.com/ansh0014/booking/service"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize database connection
	if err := config.InitMongo(); err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	log.Println("MongoDB connected successfully")

	// Initialize Redis
	if err := config.InitRedis(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	log.Println("Redis connected successfully")

	// Initialize platform services
	platformServices := initPlatformServices()
	log.Println("Platform services initialized")

	// Setup router with platform services
	r := router.SetupRoutes(platformServices)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	// Start the server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Booking service starting on port %s...", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func initPlatformServices() map[string]interface{} {
	db := config.MongoDB
	redisClient := config.RedisClient

	// Create repositories
	flightRepo := flight.NewRepository(db)
	railwayRepo := railway.NewRepository(db)
	eventRepo := event.NewRepository(db)
	movieRepo := movie.NewRepository(db)

	// Create services
	flightService := flight.NewService(flightRepo, redisClient)
	railwayService := railway.NewService(railwayRepo, redisClient)
	eventService := event.NewService(eventRepo, redisClient)
	movieService := movie.NewService(movieRepo, redisClient)

	// Create booking and seat services
	bookingService := service.NewBookingService(db, redisClient)
	seatService := service.NewSeatService(redisClient, map[string]interface{}{
		"flight":  flightService,
		"railway": railwayService,
		"event":   eventService,
		"movie":   movieService,
	})

	// Set up circular reference
	bookingService.SetSeatService(seatService)

	// Return all services in a map
	return map[string]interface{}{
		"flight":  flightService,
		"railway": railwayService,
		"event":   eventService,
		"movie":   movieService,
		"booking": bookingService,
		"seat":    seatService,
	}
}

// Graceful shutdown handler
func setupGracefulShutdown(server *http.Server) {
	// Create a channel to listen for interrupt signals
	c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-c

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	log.Println("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}
