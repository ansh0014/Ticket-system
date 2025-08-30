package config

import (
	"context"
	"os"
	"time"

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

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"), // e.g. "localhost:6379"
		Password: "",                     // or from env
		DB:       0,
	})
}

func InitMongo() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
    if err != nil {
        return err
    }

    // Verify connection
    if err := client.Ping(ctx, nil); err != nil {
        return err
    }

    MongoClient = client
    MongoDB = client.Database(os.Getenv("MONGODB_DATABASE"))
    return nil
}
