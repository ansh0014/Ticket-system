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
        Addr:     os.Getenv("REDIS_URL"),
        Password: "", // no password set
        DB:       0,  // use default DB
    })
}

func InitMongo() error {
    client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
    if err != nil {
        return err
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := client.Connect(ctx); err != nil {
        return err
    }
    MongoClient = client
    MongoDB = client.Database(os.Getenv("MONGODB_DATABASE"))
    return nil
}