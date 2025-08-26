package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ansh0014/auth/config"
	"github.com/ansh0014/auth/router"
)

func main() {
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
