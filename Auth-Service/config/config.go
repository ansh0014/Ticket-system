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
