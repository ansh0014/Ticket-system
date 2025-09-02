package railway

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Station represents a railway station
type Station struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Code      string             `json:"code" bson:"code"`
	Name      string             `json:"name" bson:"name"`
	City      string             `json:"city" bson:"city"`
	State     string             `json:"state" bson:"state"`
	Country   string             `json:"country" bson:"country"`
	Address   string             `json:"address" bson:"address"`
	ZipCode   string             `json:"zip_code" bson:"zip_code"`
	Platforms int                `json:"platforms" bson:"platforms"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// Train represents a train in the system
type Train struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Number         string             `json:"number" bson:"number"`
	Name           string             `json:"name" bson:"name"`
	Type           string             `json:"type" bson:"type"` // Express, Passenger, etc.
	OriginID       primitive.ObjectID `json:"origin_id" bson:"origin_id"`
	Origin         Station            `json:"origin" bson:"origin,omitempty"`
	DestinationID  primitive.ObjectID `json:"destination_id" bson:"destination_id"`
	Destination    Station            `json:"destination" bson:"destination,omitempty"`
	DepartureTime  time.Time          `json:"departure_time" bson:"departure_time"`
	ArrivalTime    time.Time          `json:"arrival_time" bson:"arrival_time"`
	Duration       int                `json:"duration" bson:"duration"` // in minutes
	Distance       float64            `json:"distance" bson:"distance"` // in kilometers
	TotalSeats     int                `json:"total_seats" bson:"total_seats"`
	AvailableSeats int                `json:"available_seats" bson:"available_seats"`
	Status         string             `json:"status" bson:"status"`       // scheduled, delayed, cancelled, completed
	Frequency      []string           `json:"frequency" bson:"frequency"` // days of the week
	Classes        []string           `json:"classes" bson:"classes"`     // 1A, 2A, 3A, SL, etc.
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

// TrainStop represents a stop for a train
type TrainStop struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TrainID            primitive.ObjectID `json:"train_id" bson:"train_id"`
	StationID          primitive.ObjectID `json:"station_id" bson:"station_id"`
	Station            Station            `json:"station" bson:"station,omitempty"`
	ArrivalTime        time.Time          `json:"arrival_time" bson:"arrival_time"`
	DepartureTime      time.Time          `json:"departure_time" bson:"departure_time"`
	Day                int                `json:"day" bson:"day"` // Day of journey
	StopNumber         int                `json:"stop_number" bson:"stop_number"`
	DistanceFromOrigin float64            `json:"distance_from_origin" bson:"distance_from_origin"`
	Platform           string             `json:"platform" bson:"platform"`
	CreatedAt          time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" bson:"updated_at"`
}

// TrainSeat represents a seat on a train
type TrainSeat struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TrainID     primitive.ObjectID `json:"train_id" bson:"train_id"`
	Coach       string             `json:"coach" bson:"coach"`
	SeatNumber  string             `json:"seat_number" bson:"seat_number"`
	Class       string             `json:"class" bson:"class"` // 1A, 2A, 3A, SL, etc.
	Type        string             `json:"type" bson:"type"`   // Window, Middle, Aisle, Lower, Upper, etc.
	IsAvailable bool               `json:"is_available" bson:"is_available"`
	Price       float64            `json:"price" bson:"price"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// SearchTrainsRequest represents a request to search for trains
type SearchTrainsRequest struct {
	Origin      string    `json:"origin"`
	Destination string    `json:"destination"`
	Date        time.Time `json:"date"`
	Class       string    `json:"class"`
	Passengers  int       `json:"passengers"`
}

// TrainResponse represents a train with additional formatted data
type TrainResponse struct {
	Train         *Train `json:"train"`
	Duration      string `json:"duration_formatted"`
	DepartureStr  string `json:"departure_str"`
	ArrivalStr    string `json:"arrival_str"`
	DepartureDate string `json:"departure_date"`
	ArrivalDate   string `json:"arrival_date"`
}
