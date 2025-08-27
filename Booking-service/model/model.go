package model

import (
	"time"
)

type Booking struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	UserID      string    `json:"user_id" bson:"user_id"`
	ShowID      string    `json:"show_id" bson:"show_id"`
	Seats       []string  `json:"seats" bson:"seats"`
	TotalPrice  float64   `json:"total_price" bson:"total_price"`
	Status      string    `json:"status" bson:"status"`
	PaymentID   string    `json:"payment_id,omitempty" bson:"payment_id,omitempty"`
	BookingTime time.Time `json:"booking_time" bson:"booking_time"`
	ExpiryTime  time.Time `json:"expiry_time,omitempty" bson:"expiry_time,omitempty"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

type Seat struct {
	ID       string  `json:"id" bson:"_id,omitempty"`
	ShowID   string  `json:"show_id" bson:"show_id"`
	Row      string  `json:"row" bson:"row"`
	Number   int     `json:"number" bson:"number"`
	Category string  `json:"category" bson:"category"` 
	Price    float64 `json:"price" bson:"price"`
	Status   string  `json:"status" bson:"status"` 
}


type CreateBookingRequest struct {
	UserID string   `json:"user_id" validate:"required"`
	ShowID string   `json:"show_id" validate:"required"`
	Seats  []string `json:"seats" validate:"required,min=1"`
}

type BookingResponse struct {
	Booking    Booking `json:"booking"`
	PaymentURL string  `json:"payment_url,omitempty"`
	ExpiresIn  int     `json:"expires_in,omitempty"` 
}

type LockSeatsRequest struct {
	ShowID   string   `json:"show_id" validate:"required"`
	Seats    []string `json:"seats" validate:"required,min=1"`
	UserID   string   `json:"user_id" validate:"required"`
	Duration int      `json:"duration,omitempty"` 
}

type AvailabilityRequest struct {
	ShowID string `json:"show_id" validate:"required"`
}

type AvailabilityResponse struct {
	ShowID    string    `json:"show_id"`
	Available []string  `json:"available"`
	Locked    []string  `json:"locked"`
	Booked    []string  `json:"booked"`
	UpdatedAt time.Time `json:"updated_at"`
}
