package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/ansh0014/venue/handler"
    "github.com/ansh0014/venue/repository"
    "github.com/ansh0014/venue/routes"
    "github.com/ansh0014/venue/service"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" {
        mongoURI = "mongodb://localhost:27017"
    }
    port := os.Getenv("PORT")
    if port == "" {
        port = "8004"
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
    if err != nil {
        log.Fatalf("mongo connect: %v", err)
    }
    if err := client.Ping(ctx, nil); err != nil {
        log.Fatalf("mongo ping: %v", err)
    }
    db := client.Database("venue_service")

    repo := repository.NewRepository(db)
    svc := service.NewService(repo)
    h := handler.NewHandler(svc)

    router := routes.NewRouter(h)

    srv := &http.Server{
        Addr:         ":" + port,
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 20 * time.Second,
    }

    log.Printf("venue-service listening on :%s (mongo: %s)", port, mongoURI)
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}