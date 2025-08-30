package config

import (
    "context"
    "os"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var (
    MongoClient *mongo.Client
    MongoDB     *mongo.Database
    Ctx         = context.Background()
)

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

// PaymentGateway configuration
type PaymentGatewayConfig struct {
    Name      string
    APIKey    string
    SecretKey string
    BaseURL   string
    IsTest    bool
}

func GetPaymentGatewayConfig() PaymentGatewayConfig {
    return PaymentGatewayConfig{
        Name:      os.Getenv("PAYMENT_GATEWAY_NAME"),
        APIKey:    os.Getenv("PAYMENT_GATEWAY_API_KEY"),
        SecretKey: os.Getenv("PAYMENT_GATEWAY_SECRET_KEY"),
        BaseURL:   os.Getenv("PAYMENT_GATEWAY_BASE_URL"),
        IsTest:    os.Getenv("PAYMENT_GATEWAY_MODE") != "production",
    }
}