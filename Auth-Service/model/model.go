package model

import "time"

type User struct {
    ID        string    `bson:"_id,omitempty" json:"id"`
    Email     string    `bson:"email" json:"email"`
    Password  string    `bson:"password" json:"-"`
    IsActive  bool      `bson:"is_active" json:"is_active"`
    CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type OTP struct {
    Email     string    `bson:"email" json:"email"`
    Code      string    `bson:"code" json:"code"`
    ExpiresAt time.Time `bson:"expires_at" json:"expires_at"`
}
type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}