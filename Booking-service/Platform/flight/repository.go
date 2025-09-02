package flight

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository handles flight data access
type Repository struct {
	db           *mongo.Database
	flightsColl  *mongo.Collection
	airportsColl *mongo.Collection
	airlinesColl *mongo.Collection
	seatsColl    *mongo.Collection
}

// NewRepository creates a new flight repository
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		db:           db,
		flightsColl:  db.Collection("flights"),
		airportsColl: db.Collection("airports"),
		airlinesColl: db.Collection("airlines"),
		seatsColl:    db.Collection("flight_seats"),
	}
}

// GetAirportByCode retrieves an airport by its code
func (r *Repository) GetAirportByCode(ctx context.Context, code string) (*Airport, error) {
	var airport Airport
	err := r.airportsColl.FindOne(ctx, bson.M{"code": code}).Decode(&airport)
	if err != nil {
		return nil, err
	}
	return &airport, nil
}

// SearchFlights searches for flights based on criteria
func (r *Repository) SearchFlights(ctx context.Context, search SearchFlightsRequest) ([]Flight, error) {
	// Get origin and destination airports first
	originAirport, err := r.GetAirportByCode(ctx, search.Origin)
	if err != nil {
		return nil, errors.New("origin airport not found")
	}

	destAirport, err := r.GetAirportByCode(ctx, search.Destination)
	if err != nil {
		return nil, errors.New("destination airport not found")
	}

	// Create a time range for the departure date (full day)
	startOfDay := time.Date(search.DepartureDate.Year(), search.DepartureDate.Month(), search.DepartureDate.Day(), 0, 0, 0, 0, search.DepartureDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Build query filter
	filter := bson.M{
		"origin_id":      originAirport.ID,
		"destination_id": destAirport.ID,
		"departure_time": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
		"available_seats": bson.M{"$gte": search.Passengers},
		"status":          "scheduled",
	}

	// Add class-specific filtering if needed
	if search.Class != "" {
		switch search.Class {
		case "economy":
			filter["price_economy"] = bson.M{"$gt": 0}
		case "business":
			filter["price_business"] = bson.M{"$gt": 0}
		case "first":
			filter["price_first_class"] = bson.M{"$gt": 0}
		}
	}

	// Set up options for sorting
	opts := options.Find().SetSort(bson.D{{"departure_time", 1}})

	// Execute the query
	cursor, err := r.flightsColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var flights []Flight
	if err = cursor.All(ctx, &flights); err != nil {
		return nil, err
	}

	// Populate related data for each flight
	for i := range flights {
		// Populate airline
		airline, err := r.getAirlineByID(ctx, flights[i].AirlineID)
		if err == nil {
			flights[i].Airline = *airline
		}

		// Populate origin
		origin, err := r.getAirportByID(ctx, flights[i].OriginID)
		if err == nil {
			flights[i].Origin = *origin
		}

		// Populate destination
		destination, err := r.getAirportByID(ctx, flights[i].DestinationID)
		if err == nil {
			flights[i].Destination = *destination
		}
	}

	return flights, nil
}

// GetFlightSeats retrieves all seats for a flight
func (r *Repository) GetFlightSeats(ctx context.Context, flightID primitive.ObjectID) ([]FlightSeat, error) {
	filter := bson.M{"flight_id": flightID}

	cursor, err := r.seatsColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var seats []FlightSeat
	if err = cursor.All(ctx, &seats); err != nil {
		return nil, err
	}

	return seats, nil
}

// LockFlightSeats locks seats for a booking
func (r *Repository) LockFlightSeats(ctx context.Context, flightID primitive.ObjectID, seatIDs []primitive.ObjectID) error {
	filter := bson.M{
		"flight_id":    flightID,
		"_id":          bson.M{"$in": seatIDs},
		"is_available": true,
	}

	update := bson.M{
		"$set": bson.M{
			"is_available": false,
			"updated_at":   time.Now(),
		},
	}

	result, err := r.seatsColl.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	// Check if all seats were updated
	if result.ModifiedCount != int64(len(seatIDs)) {
		// Revert the changes
		revertFilter := bson.M{
			"flight_id": flightID,
			"_id":       bson.M{"$in": seatIDs},
		}

		revertUpdate := bson.M{
			"$set": bson.M{
				"is_available": true,
				"updated_at":   time.Now(),
			},
		}

		_, _ = r.seatsColl.UpdateMany(ctx, revertFilter, revertUpdate)

		return errors.New("some seats are no longer available")
	}

	// Update available seats count on the flight
	_, err = r.flightsColl.UpdateOne(
		ctx,
		bson.M{"_id": flightID},
		bson.M{
			"$inc": bson.M{"available_seats": -len(seatIDs)},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)

	return err
}

// GetFlightByID retrieves a flight by ID
func (r *Repository) GetFlightByID(ctx context.Context, id primitive.ObjectID) (*Flight, error) {
	var flight Flight
	err := r.flightsColl.FindOne(ctx, bson.M{"_id": id}).Decode(&flight)
	if err != nil {
		return nil, err
	}

	// Populate airline information
	airline, err := r.getAirlineByID(ctx, flight.AirlineID)
	if err == nil {
		flight.Airline = *airline
	}

	// Populate origin airport
	origin, err := r.getAirportByID(ctx, flight.OriginID)
	if err == nil {
		flight.Origin = *origin
	}

	// Populate destination airport
	destination, err := r.getAirportByID(ctx, flight.DestinationID)
	if err == nil {
		flight.Destination = *destination
	}

	return &flight, nil
}

// Helper methods to populate related data
func (r *Repository) getAirlineByID(ctx context.Context, id primitive.ObjectID) (*Airline, error) {
	var airline Airline
	err := r.airlinesColl.FindOne(ctx, bson.M{"_id": id}).Decode(&airline)
	if err != nil {
		return nil, err
	}
	return &airline, nil
}

func (r *Repository) getAirportByID(ctx context.Context, id primitive.ObjectID) (*Airport, error) {
	var airport Airport
	err := r.airportsColl.FindOne(ctx, bson.M{"_id": id}).Decode(&airport)
	if err != nil {
		return nil, err
	}
	return &airport, nil
}

// UnlockFlightSeats unlocks previously locked seats
func (r *Repository) UnlockFlightSeats(ctx context.Context, flightID primitive.ObjectID, seatIDs []primitive.ObjectID) error {
	filter := bson.M{
		"flight_id": flightID,
		"_id":       bson.M{"$in": seatIDs},
	}

	update := bson.M{
		"$set": bson.M{
			"is_available": true,
			"updated_at":   time.Now(),
		},
	}

	_, err := r.seatsColl.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	// Update available seats count on the flight
	_, err = r.flightsColl.UpdateOne(
		ctx,
		bson.M{"_id": flightID},
		bson.M{
			"$inc": bson.M{"available_seats": len(seatIDs)},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)

	return err
}

// GetSeatsByIDs retrieves seats by their IDs
func (r *Repository) GetSeatsByIDs(ctx context.Context, flightID primitive.ObjectID, seatIDs []primitive.ObjectID) ([]FlightSeat, error) {
	filter := bson.M{
		"flight_id": flightID,
		"_id":       bson.M{"$in": seatIDs},
	}

	cursor, err := r.seatsColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var seats []FlightSeat
	if err = cursor.All(ctx, &seats); err != nil {
		return nil, err
	}

	return seats, nil
}

// CreateFlight creates a new flight (admin function)
func (r *Repository) CreateFlight(ctx context.Context, flight *Flight) error {
	flight.CreatedAt = time.Now()
	flight.UpdatedAt = time.Now()

	result, err := r.flightsColl.InsertOne(ctx, flight)
	if err != nil {
		return err
	}

	flight.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// UpdateFlight updates an existing flight (admin function)
func (r *Repository) UpdateFlight(ctx context.Context, flight *Flight) error {
	flight.UpdatedAt = time.Now()

	_, err := r.flightsColl.ReplaceOne(
		ctx,
		bson.M{"_id": flight.ID},
		flight,
	)

	return err
}
