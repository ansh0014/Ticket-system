package model

import (
    "time"
)

type PaymentStatus string

const (
    PaymentStatusPending   PaymentStatus = "pending"
    PaymentStatusCompleted PaymentStatus = "completed"
    PaymentStatusFailed    PaymentStatus = "failed"
    PaymentStatusRefunded  PaymentStatus = "refunded"
    PaymentStatusCancelled PaymentStatus = "cancelled"
)

type PaymentMethod string

const (
    PaymentMethodCard       PaymentMethod = "card"
    PaymentMethodUPI        PaymentMethod = "upi"
    PaymentMethodNetBanking PaymentMethod = "netbanking"
    PaymentMethodWallet     PaymentMethod = "wallet"
)

type Payment struct {
    ID               string         `json:"id" bson:"_id,omitempty"`
    BookingID        string         `json:"booking_id" bson:"booking_id"`
    UserID           string         `json:"user_id" bson:"user_id"`
    Amount           float64        `json:"amount" bson:"amount"`
    Currency         string         `json:"currency" bson:"currency"`
    Status           PaymentStatus  `json:"status" bson:"status"`
    Method           PaymentMethod  `json:"method,omitempty" bson:"method,omitempty"`
    GatewayReference string         `json:"gateway_reference,omitempty" bson:"gateway_reference,omitempty"`
    PaymentURL       string         `json:"payment_url,omitempty" bson:"payment_url,omitempty"`
    CallbackURL      string         `json:"callback_url,omitempty" bson:"callback_url,omitempty"`
    CreatedAt        time.Time      `json:"created_at" bson:"created_at"`
    UpdatedAt        time.Time      `json:"updated_at" bson:"updated_at"`
    CompletedAt      *time.Time     `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
}

type Transaction struct {
    ID           string    `json:"id" bson:"_id,omitempty"`
    PaymentID    string    `json:"payment_id" bson:"payment_id"`
    Type         string    `json:"type" bson:"type"` // "charge", "refund", "cancel"
    Amount       float64   `json:"amount" bson:"amount"`
    Status       string    `json:"status" bson:"status"`
    ErrorMessage string    `json:"error_message,omitempty" bson:"error_message,omitempty"`
    Metadata     string    `json:"metadata,omitempty" bson:"metadata,omitempty"`
    CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}

// Request/Response Types
type CreatePaymentRequest struct {
    BookingID   string  `json:"booking_id" validate:"required"`
    UserID      string  `json:"user_id" validate:"required"`
    Amount      float64 `json:"amount" validate:"required,gt=0"`
    Currency    string  `json:"currency" validate:"required"`
    CallbackURL string  `json:"callback_url"`
}

type PaymentResponse struct {
    Payment     Payment `json:"payment"`
    RedirectURL string  `json:"redirect_url,omitempty"`
    ExpiresIn   int     `json:"expires_in,omitempty"` // Seconds until payment expires
}

type WebhookRequest struct {
    GatewayReference string         `json:"gateway_reference"`
    Status           PaymentStatus  `json:"status"`
    Amount           float64        `json:"amount"`
    Currency         string         `json:"currency"`
    Method           PaymentMethod  `json:"method"`
    Metadata         string         `json:"metadata,omitempty"`
}

type VerifyPaymentRequest struct {
    PaymentID string `json:"payment_id" validate:"required"`
}

type RefundRequest struct {
    PaymentID string  `json:"payment_id" validate:"required"`
    Amount    float64 `json:"amount" validate:"required,gt=0"`
    Reason    string  `json:"reason"`
}