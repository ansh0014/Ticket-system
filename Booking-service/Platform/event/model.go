package event

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

// Venue represents an event venue
type Venue struct {
    ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Name      string             `json:"name" bson:"name"`
    Address   string             `json:"address" bson:"address"`
    City      string             `json:"city" bson:"city"`
    State     string             `json:"state" bson:"state"`
    Country   string             `json:"country" bson:"country"`
    ZipCode   string             `json:"zip_code" bson:"zip_code"`
    Capacity  int                `json:"capacity" bson:"capacity"`
    Latitude  float64            `json:"latitude" bson:"latitude"`
    Longitude float64            `json:"longitude" bson:"longitude"`
    CreatedAt time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// Organizer represents an event organizer
type Organizer struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Name        string             `json:"name" bson:"name"`
    Description string             `json:"description" bson:"description"`
    Logo        string             `json:"logo" bson:"logo"`
    Website     string             `json:"website" bson:"website"`
    Email       string             `json:"email" bson:"email"`
    Phone       string             `json:"phone" bson:"phone"`
    CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Category represents an event category
type Category struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Name        string             `json:"name" bson:"name"`
    Description string             `json:"description" bson:"description"`
    CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Event represents an event in the system
type Event struct {
    ID             primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
    Title          string               `json:"title" bson:"title"`
    Description    string               `json:"description" bson:"description"`
    BannerImage    string               `json:"banner_image" bson:"banner_image"`
    Images         []string             `json:"images" bson:"images"`
    StartTime      time.Time            `json:"start_time" bson:"start_time"`
    EndTime        time.Time            `json:"end_time" bson:"end_time"`
    VenueID        primitive.ObjectID   `json:"venue_id" bson:"venue_id"`
    Venue          Venue                `json:"venue" bson:"venue,omitempty"`
    OrganizerID    primitive.ObjectID   `json:"organizer_id" bson:"organizer_id"`
    Organizer      Organizer            `json:"organizer" bson:"organizer,omitempty"`
    CategoryIDs    []primitive.ObjectID `json:"category_ids" bson:"category_ids"`
    Categories     []Category           `json:"categories" bson:"categories,omitempty"`
    Status         string               `json:"status" bson:"status"` // upcoming, ongoing, completed, cancelled
    TicketTypes    []TicketType         `json:"ticket_types" bson:"ticket_types"`
    TotalCapacity  int                  `json:"total_capacity" bson:"total_capacity"`
    AvailableSeats int                  `json:"available_seats" bson:"available_seats"`
    CreatedAt      time.Time            `json:"created_at" bson:"created_at"`
    UpdatedAt      time.Time            `json:"updated_at" bson:"updated_at"`
}

// TicketType represents a type of ticket for an event
type TicketType struct {
    ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    EventID         primitive.ObjectID `json:"event_id" bson:"event_id"`
    Name            string             `json:"name" bson:"name"`
    Description     string             `json:"description" bson:"description"`
    Price           float64            `json:"price" bson:"price"`
    Currency        string             `json:"currency" bson:"currency"`
    Quantity        int                `json:"quantity" bson:"quantity"`
    AvailableCount  int                `json:"available_count" bson:"available_count"`
    MaxPerCustomer  int                `json:"max_per_customer" bson:"max_per_customer"`
    SaleStartTime   time.Time          `json:"sale_start_time" bson:"sale_start_time"`
    SaleEndTime     time.Time          `json:"sale_end_time" bson:"sale_end_time"`
    HasReservedSeat bool               `json:"has_reserved_seat" bson:"has_reserved_seat"`
    CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

// EventSeat represents a seat for an event
type EventSeat struct {
    ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    EventID      primitive.ObjectID `json:"event_id" bson:"event_id"`
    TicketTypeID primitive.ObjectID `json:"ticket_type_id" bson:"ticket_type_id"`
    SeatNumber   string             `json:"seat_number" bson:"seat_number"`
    Row          string             `json:"row" bson:"row"`
    Section      string             `json:"section" bson:"section"`
    IsAvailable  bool               `json:"is_available" bson:"is_available"`
    Price        float64            `json:"price" bson:"price"`
    CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// SearchEventsRequest represents a request to search for events
type SearchEventsRequest struct {
    Query       string    `json:"query"`
    City        string    `json:"city"`
    Category    string    `json:"category"`
    StartDate   time.Time `json:"start_date"`
    EndDate     time.Time `json:"end_date"`
    PriceMin    float64   `json:"price_min"`
    PriceMax    float64   `json:"price_max"`
    TicketCount int       `json:"ticket_count"`
}

// EventResponse represents the response for an event search or detail
type EventResponse struct {
    Event       *Event
    DateDisplay string
    TimeDisplay string
    DayOfWeek   string
    Month       string
}