package flight

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Airport represents an airport in the system
type Airport struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Code      string             `json:"code" bson:"code"`
	Name      string             `json:"name" bson:"name"`
	City      string             `json:"city" bson:"city"`
	Country   string             `json:"country" bson:"country"`
	Terminal  string             `json:"terminal" bson:"terminal"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// Airline represents an airline company
type Airline struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Code      string             `json:"code" bson:"code"`
	Name      string             `json:"name" bson:"name"`
	Logo      string             `json:"logo" bson:"logo"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// Flight represents a flight in the system
type Flight struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FlightNumber    string             `json:"flight_number" bson:"flight_number"`
	AirlineID       primitive.ObjectID `json:"airline_id" bson:"airline_id"`
	Airline         Airline            `json:"airline" bson:"airline,omitempty"`
	DepartureTime   time.Time          `json:"departure_time" bson:"departure_time"`
	ArrivalTime     time.Time          `json:"arrival_time" bson:"arrival_time"`
	OriginID        primitive.ObjectID `json:"origin_id" bson:"origin_id"`
	Origin          Airport            `json:"origin" bson:"origin,omitempty"`
	DestinationID   primitive.ObjectID `json:"destination_id" bson:"destination_id"`
	Destination     Airport            `json:"destination" bson:"destination,omitempty"`
	Duration        int                `json:"duration" bson:"duration"` // in minutes
	Status          string             `json:"status" bson:"status"`     // scheduled, delayed, cancelled, completed
	Aircraft        string             `json:"aircraft" bson:"aircraft"`
	TotalSeats      int                `json:"total_seats" bson:"total_seats"`
	AvailableSeats  int                `json:"available_seats" bson:"available_seats"`
	PriceEconomy    float64            `json:"price_economy" bson:"price_economy"`
	PriceBusiness   float64            `json:"price_business" bson:"price_business"`
	PriceFirstClass float64            `json:"price_first_class" bson:"price_first_class"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

// FlightSeat represents a seat on a flight
type FlightSeat struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FlightID    primitive.ObjectID `json:"flight_id" bson:"flight_id"`
	SeatNumber  string             `json:"seat_number" bson:"seat_number"`
	Class       string             `json:"class" bson:"class"` // economy, business, first
	IsAvailable bool               `json:"is_available" bson:"is_available"`
	Price       float64            `json:"price" bson:"price"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// SearchFlightsRequest represents a flight search request
type SearchFlightsRequest struct {
	Origin        string    `json:"origin"`
	Destination   string    `json:"destination"`
	DepartureDate time.Time `json:"departure_date"`
	ReturnDate    time.Time `json:"return_date,omitempty"`
	Passengers    int       `json:"passengers"`
	Class         string    `json:"class"`
}

