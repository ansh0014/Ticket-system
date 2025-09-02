package flight

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FlightResponse represents the response for a flight search or detail
type FlightResponse struct {
	Flight       *Flight
	Duration     string
	DepartureDay string
	DepartureStr string
	ArrivalStr   string
	Price        float64
}

// Service handles flight business logic
type Service struct {
	repo         *Repository
	redisClient  *redis.Client
	seatLockTime time.Duration
}

// NewService creates a new flight service
func NewService(repo *Repository, redisClient *redis.Client) *Service {
	return &Service{
		repo:         repo,
		redisClient:  redisClient,
		seatLockTime: 5 * time.Minute, // Default 5 minutes lock time
	}
}

// SearchFlights searches for flights matching the criteria
func (s *Service) SearchFlights(ctx context.Context, req SearchFlightsRequest) ([]FlightResponse, error) {
	// Validate the search request
	if req.Origin == "" || req.Destination == "" {
		return nil, errors.New("origin and destination are required")
	}

	if req.Origin == req.Destination {
		return nil, errors.New("origin and destination cannot be the same")
	}

	if req.Passengers <= 0 {
		return nil, errors.New("at least one passenger is required")
	}

	// Set default class if not specified
	if req.Class == "" {
		req.Class = "economy"
	}

	// Ensure date is valid
	if req.DepartureDate.Before(time.Now().Add(-24 * time.Hour)) {
		return nil, errors.New("departure date must be in the future")
	}

	// Perform the search
	flights, err := s.repo.SearchFlights(ctx, req)
	if err != nil {
		return nil, err
	}

	// Format response
	var response []FlightResponse
	for _, flight := range flights {
		// Skip flight if there aren't enough seats
		if flight.AvailableSeats < req.Passengers {
			continue
		}

		// Calculate price based on class
		var price float64
		switch req.Class {
		case "economy":
			price = flight.PriceEconomy
		case "business":
			price = flight.PriceBusiness
		case "first":
			price = flight.PriceFirstClass
		default:
			price = flight.PriceEconomy
		}

		// Format the flight data
		flightResp := FlightResponse{
			Flight:       &flight,
			Duration:     formatDuration(flight.Duration),
			DepartureDay: flight.DepartureTime.Format("Mon, 02 Jan"),
			DepartureStr: flight.DepartureTime.Format("15:04"),
			ArrivalStr:   flight.ArrivalTime.Format("15:04"),
			Price:        price,
		}

		response = append(response, flightResp)
	}

	return response, nil
}

// GetFlightByID gets detailed information about a flight
func (s *Service) GetFlightByID(ctx context.Context, flightID string) (*FlightResponse, error) {
	id, err := primitive.ObjectIDFromHex(flightID)
	if err != nil {
		return nil, errors.New("invalid flight ID")
	}

	flight, err := s.repo.GetFlightByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Format the flight data
	response := &FlightResponse{
		Flight:       flight,
		Duration:     formatDuration(flight.Duration),
		DepartureDay: flight.DepartureTime.Format("Mon, 02 Jan"),
		DepartureStr: flight.DepartureTime.Format("15:04"),
		ArrivalStr:   flight.ArrivalTime.Format("15:04"),
	}

	return response, nil
}

// GetFlightSeats retrieves the seat map for a flight
func (s *Service) GetFlightSeats(ctx context.Context, flightID string) ([]FlightSeat, error) {
	id, err := primitive.ObjectIDFromHex(flightID)
	if err != nil {
		return nil, errors.New("invalid flight ID")
	}

	return s.repo.GetFlightSeats(ctx, id)
}

// LockFlightSeats temporarily locks seats for a booking
func (s *Service) LockFlightSeats(ctx context.Context, flightID string, seatIDs []string, userID string) error {
	if flightID == "" {
		return errors.New("flight ID is required")
	}

	if len(seatIDs) == 0 {
		return errors.New("at least one seat must be selected")
	}

	if userID == "" {
		return errors.New("user ID is required")
	}

	// Convert string IDs to ObjectIDs
	if _, err := primitive.ObjectIDFromHex(flightID); err != nil {
		return errors.New("invalid flight ID")
	}

	seatObjIDs := make([]primitive.ObjectID, 0, len(seatIDs))
	for _, id := range seatIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return errors.New("invalid seat ID: " + id)
		}
		seatObjIDs = append(seatObjIDs, objID)
	}

	// Check if seats are already locked in Redis
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("flight_seat_lock:%s:%s", flightID, seatID)

		// Check if the lock exists
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val != "" && val != userID {
			// Seat is locked by someone else
			return errors.New("one or more selected seats are no longer available")
		}
	}

	// Set Redis locks
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("flight_seat_lock:%s:%s", flightID, seatID)

		// Set the lock with expiration
		err := s.redisClient.Set(ctx, lockKey, userID, s.seatLockTime).Err()
		if err != nil {
			// If any lock fails, clean up and return error
			s.unlockSeats(ctx, flightID, seatIDs, userID)
			return errors.New("failed to lock seats: " + err.Error())
		}
	}

	return nil
}

// unlockSeats helper to remove Redis locks
func (s *Service) unlockSeats(ctx context.Context, flightID string, seatIDs []string, userID string) {
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("flight_seat_lock:%s:%s", flightID, seatID)

		// Only remove if the lock belongs to this user
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val == userID {
			s.redisClient.Del(ctx, lockKey)
		}
	}
}

// ConfirmSeats permanently reserves seats after payment
func (s *Service) ConfirmSeats(ctx context.Context, flightID string, seatIDs []string) error {
	// Convert string IDs to ObjectIDs
	flightObjID, err := primitive.ObjectIDFromHex(flightID)
	if err != nil {
		return errors.New("invalid flight ID")
	}

	seatObjIDs := make([]primitive.ObjectID, 0, len(seatIDs))
	for _, id := range seatIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return errors.New("invalid seat ID: " + id)
		}
		seatObjIDs = append(seatObjIDs, objID)
	}

	// Remove Redis locks
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("flight_seat_lock:%s:%s", flightID, seatID)
		s.redisClient.Del(ctx, lockKey)
	}

	// Update the database
	return s.repo.LockFlightSeats(ctx, flightObjID, seatObjIDs)
}

// ReleaseSeats releases previously locked seats
func (s *Service) ReleaseSeats(ctx context.Context, flightID string, seatIDs []string, userID string) error {
	// Convert string IDs to ObjectIDs
	flightObjID, err := primitive.ObjectIDFromHex(flightID)
	if err != nil {
		return errors.New("invalid flight ID")
	}

	seatObjIDs := make([]primitive.ObjectID, 0, len(seatIDs))
	for _, id := range seatIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return errors.New("invalid seat ID: " + id)
		}
		seatObjIDs = append(seatObjIDs, objID)
	}

	// Remove Redis locks
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("flight_seat_lock:%s:%s", flightID, seatID)

		// Only remove if the lock belongs to this user
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val == userID {
			s.redisClient.Del(ctx, lockKey)
		}
	}

	// Update the database if these were permanent locks
	return s.repo.UnlockFlightSeats(ctx, flightObjID, seatObjIDs)
}

// Helper function to format duration in minutes to a readable string
func formatDuration(minutes int) string {
	hours := minutes / 60
	mins := minutes % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
