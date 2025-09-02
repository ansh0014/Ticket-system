package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles event business logic
type Service struct {
	repo         *Repository
	redisClient  *redis.Client
	seatLockTime time.Duration
}

// NewService creates a new event service
func NewService(repo *Repository, redisClient *redis.Client) *Service {
	return &Service{
		repo:         repo,
		redisClient:  redisClient,
		seatLockTime: 5 * time.Minute, // Default 5 minutes lock time
	}
}

// SearchEvents searches for events matching the criteria
func (s *Service) SearchEvents(ctx context.Context, req SearchEventsRequest, page, pageSize int) ([]EventResponse, int64, error) {
	// Validate the search request
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Perform the search
	events, total, err := s.repo.SearchEvents(ctx, req, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// Format response
	var response []EventResponse
	for _, event := range events {
		// Format the event data
		eventResp := EventResponse{
			Event:       &event,
			DateDisplay: event.StartTime.Format("Mon, 02 Jan 2006"),
			TimeDisplay: event.StartTime.Format("15:04"),
			DayOfWeek:   event.StartTime.Format("Mon"),
			Month:       event.StartTime.Format("Jan"),
		}

		response = append(response, eventResp)
	}

	return response, total, nil
}

// GetEventByID gets detailed information about an event
func (s *Service) GetEventByID(ctx context.Context, eventID string) (*EventResponse, error) {
	id, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return nil, errors.New("invalid event ID")
	}

	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Format the event data
	response := &EventResponse{
		Event:       event,
		DateDisplay: event.StartTime.Format("Mon, 02 Jan 2006"),
		TimeDisplay: event.StartTime.Format("15:04"),
		DayOfWeek:   event.StartTime.Format("Mon"),
		Month:       event.StartTime.Format("Jan"),
	}

	return response, nil
}

// GetEventSeats retrieves the seats for an event
func (s *Service) GetEventSeats(ctx context.Context, eventID string, ticketTypeID string) ([]EventSeat, error) {
	id, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return nil, errors.New("invalid event ID")
	}

	var ticketTypeObjID primitive.ObjectID
	if ticketTypeID != "" {
		ticketTypeObjID, err = primitive.ObjectIDFromHex(ticketTypeID)
		if err != nil {
			return nil, errors.New("invalid ticket type ID")
		}
	}

	return s.repo.GetEventSeats(ctx, id, ticketTypeObjID)
}

// GetTicketTypes retrieves the ticket types for an event
func (s *Service) GetTicketTypes(ctx context.Context, eventID string) ([]TicketType, error) {
	id, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return nil, errors.New("invalid event ID")
	}

	return s.repo.GetTicketTypes(ctx, id)
}

// LockEventSeats temporarily locks seats for a booking
func (s *Service) LockEventSeats(ctx context.Context, eventID string, ticketTypeID string, seatIDs []string, userID string) error {
	if eventID == "" {
		return errors.New("event ID is required")
	}

	if len(seatIDs) == 0 {
		return errors.New("at least one seat must be selected")
	}

	if userID == "" {
		return errors.New("user ID is required")
	}

	// Validate event ID
	_, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return errors.New("invalid event ID")
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
		lockKey := fmt.Sprintf("event_seat_lock:%s:%s", eventID, seatID)

		// Check if the lock exists
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val != "" && val != userID {
			// Seat is locked by someone else
			return errors.New("one or more selected seats are no longer available")
		}
	}

	// Set Redis locks
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("event_seat_lock:%s:%s", eventID, seatID)

		// Set the lock with expiration
		err := s.redisClient.Set(ctx, lockKey, userID, s.seatLockTime).Err()
		if err != nil {
			// If any lock fails, clean up and return error
			s.unlockSeats(ctx, eventID, seatIDs, userID)
			return errors.New("failed to lock seats: " + err.Error())
		}
	}

	return nil
}

// unlockSeats helper to remove Redis locks
func (s *Service) unlockSeats(ctx context.Context, eventID string, seatIDs []string, userID string) {
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("event_seat_lock:%s:%s", eventID, seatID)

		// Only remove if the lock belongs to this user
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val == userID {
			s.redisClient.Del(ctx, lockKey)
		}
	}
}

// ConfirmSeats permanently reserves seats after payment
func (s *Service) ConfirmSeats(ctx context.Context, eventID string, ticketTypeID string, seatIDs []string) error {
	// Convert string IDs to ObjectIDs
	eventObjID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return errors.New("invalid event ID")
	}

	var ticketTypeObjID primitive.ObjectID
	if ticketTypeID != "" {
		ticketTypeObjID, err = primitive.ObjectIDFromHex(ticketTypeID)
		if err != nil {
			return errors.New("invalid ticket type ID")
		}
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
		lockKey := fmt.Sprintf("event_seat_lock:%s:%s", eventID, seatID)
		s.redisClient.Del(ctx, lockKey)
	}

	// Update the database
	return s.repo.LockEventSeats(ctx, eventObjID, ticketTypeObjID, seatObjIDs)
}

// ReleaseSeats releases previously locked seats
func (s *Service) ReleaseSeats(ctx context.Context, eventID string, seatIDs []string, userID string) error {
	// Convert string IDs to ObjectIDs
	eventObjID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return errors.New("invalid event ID")
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
		lockKey := fmt.Sprintf("event_seat_lock:%s:%s", eventID, seatID)

		// Only remove if the lock belongs to this user
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val == userID {
			s.redisClient.Del(ctx, lockKey)
		}
	}

	// Update the database if these were permanent locks
	return s.repo.UnlockEventSeats(ctx, eventObjID, seatObjIDs)
}
