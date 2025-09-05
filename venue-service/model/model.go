package model

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Venue struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Name      string             `bson:"name" json:"name"`
    Address   string             `bson:"address" json:"address"`
    City      string             `bson:"city" json:"city"`
    Meta      map[string]string  `bson:"meta,omitempty" json:"meta,omitempty"`
    CreatedAt time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type Hall struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    VenueID   primitive.ObjectID `bson:"venue_id" json:"venue_id"`
    Name      string             `bson:"name" json:"name"`
    Rows      int                `bson:"rows" json:"rows"`
    Cols      int                `bson:"cols" json:"cols"`
    SeatMap   [][]string         `bson:"seat_map,omitempty" json:"seat_map,omitempty"`
    CreatedAt time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type Seat struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    HallID    primitive.ObjectID `bson:"hall_id" json:"hall_id"`
    Number    string             `bson:"number" json:"number"`
    Category  string             `bson:"category" json:"category"`
    Price     float64            `bson:"price" json:"price"`
    IsActive  bool               `bson:"is_active" json:"is_active"`
    CreatedAt time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}