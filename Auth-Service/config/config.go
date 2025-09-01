package config

import (
	"context"
	"os"
	"time"
	"fmt"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	RedisClient *redis.Client
	MongoClient *mongo.Client
	MongoDB     *mongo.Database
	Ctx         = context.Background()
)

func InitRedis() error {
    // Parse Redis URL which includes all connection information
    opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
    if err != nil {
        return err
    }
    
    RedisClient = redis.NewClient(opt)
    
    // Ping Redis to verify connection
    _, err = RedisClient.Ping(Ctx).Result()
    return err
}
func InitMongo() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    mongoURI := os.Getenv("MONGODB_URI")
    fmt.Println("MongoDB URI:", mongoURI) // Debug print
    
    if mongoURI == "" {
        return fmt.Errorf("MONGODB_URI environment variable is not set")
    }

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
    if err != nil {
        return err
    }

    // Verify connection
    if err := client.Ping(ctx, nil); err != nil {
        return err
    }

    dbName := os.Getenv("MONGODB_DATABASE")
    fmt.Println("MongoDB Database:", dbName) // Debug print
    
    if dbName == "" {
        dbName = "authdb" // Default fallback
    }

    MongoClient = client
    MongoDB = client.Database(dbName)
    return nil
}