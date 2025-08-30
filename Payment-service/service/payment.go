package service

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/ansh0014/payment/config"
    "github.com/ansh0014/payment/model"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

// CreatePayment initializes a new payment
func CreatePayment(req *model.CreatePaymentRequest) (*model.Payment, error) {
    // Create a new payment record
    payment := &model.Payment{
        ID:          primitive.NewObjectID().Hex(),
        BookingID:   req.BookingID,
        UserID:      req.UserID,
        Amount:      req.Amount,
        Currency:    req.Currency,
        Status:      model.PaymentStatusPending,
        CallbackURL: req.CallbackURL,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    // Generate payment URL or gateway reference using payment provider
    gatewayRef, paymentURL, err := createPaymentWithProvider(payment)
    if err != nil {
        return nil, err
    }

    payment.GatewayReference = gatewayRef
    payment.PaymentURL = paymentURL

    // Insert payment into MongoDB
    _, err = config.MongoDB.Collection("payments").InsertOne(context.Background(), payment)
    if err != nil {
        return nil, err
    }

    return payment, nil
}

// GetPayment retrieves a payment by ID
func GetPayment(paymentID string) (*model.Payment, error) {
    var payment model.Payment
    err := config.MongoDB.Collection("payments").FindOne(
        context.Background(),
        bson.M{"_id": paymentID},
    ).Decode(&payment)

    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, errors.New("payment not found")
        }
        return nil, err
    }
    return &payment, nil
}

// UpdatePaymentStatus updates the status of a payment
func UpdatePaymentStatus(paymentID string, status model.PaymentStatus) error {
    update := bson.M{
        "$set": bson.M{
            "status":     status,
            "updated_at": time.Now(),
        },
    }

    // If completed, set the completed_at timestamp
    if status == model.PaymentStatusCompleted {
        now := time.Now()
        update["$set"].(bson.M)["completed_at"] = now
    }

    _, err := config.MongoDB.Collection("payments").UpdateOne(
        context.Background(),
        bson.M{"_id": paymentID},
        update,
    )

    if err != nil {
        return err
    }

    // If payment completed or failed, notify the booking service
    if status == model.PaymentStatusCompleted || status == model.PaymentStatusFailed {
        payment, err := GetPayment(paymentID)
        if err != nil {
            return err
        }
        notifyBookingService(payment, string(status))
    }

    return nil
}

// ProcessWebhook processes payment gateway webhook events
func ProcessWebhook(webhook *model.WebhookRequest) error {
    // Find payment by gateway reference
    var payment model.Payment
    err := config.MongoDB.Collection("payments").FindOne(
        context.Background(),
        bson.M{"gateway_reference": webhook.GatewayReference},
    ).Decode(&payment)

    if err != nil {
        return err
    }

    // Update payment status
    return UpdatePaymentStatus(payment.ID, webhook.Status)
}

// Refund a payment (fully or partially)
func RefundPayment(req *model.RefundRequest) error {
    payment, err := GetPayment(req.PaymentID)
    if err != nil {
        return err
    }

    if payment.Status != model.PaymentStatusCompleted {
        return errors.New("only completed payments can be refunded")
    }

    if req.Amount > payment.Amount {
        return errors.New("refund amount cannot exceed payment amount")
    }

    // Process refund with payment provider
    refundRef, err := refundPaymentWithProvider(payment, req.Amount, req.Reason)
    if err != nil {
        return err
    }

    // Create transaction record
    transaction := model.Transaction{
        ID:        primitive.NewObjectID().Hex(),
        PaymentID: payment.ID,
        Type:      "refund",
        Amount:    req.Amount,
        Status:    "success",
        Metadata:  refundRef,
        CreatedAt: time.Now(),
    }

    _, err = config.MongoDB.Collection("transactions").InsertOne(context.Background(), transaction)
    if err != nil {
        return err
    }

    // Update payment status if full refund
    if req.Amount == payment.Amount {
        return UpdatePaymentStatus(payment.ID, model.PaymentStatusRefunded)
    }

    return nil
}

// Helper function to mock payment provider integration (to be implemented)
func createPaymentWithProvider(payment *model.Payment) (string, string, error) {
    // In real implementation, this would call the payment gateway API
    // For now, we'll return mock values
    gatewayRef := fmt.Sprintf("PAY_%s", payment.ID)
    paymentURL := fmt.Sprintf("http://mock-payment-gateway.com/pay/%s", gatewayRef)
    return gatewayRef, paymentURL, nil
}

// Helper function to mock refund (to be implemented)
func refundPaymentWithProvider(payment *model.Payment, amount float64, reason string) (string, error) {
    // In real implementation, this would call the payment gateway API
    // For now, we'll return a mock reference
    return fmt.Sprintf("REF_%s", payment.ID), nil
}

// Helper function to notify the booking service (to be implemented)
func notifyBookingService(payment *model.Payment, status string) error {
    // In real implementation, this would call the booking service API
    // For now, we'll just log it
    fmt.Printf("Notifying booking service: Payment %s for booking %s is %s\n",
        payment.ID, payment.BookingID, status)
    return nil
}