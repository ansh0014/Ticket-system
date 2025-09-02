package railway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles railway business logic
type Service struct {
	repo         *Repository
	redisClient  *redis.Client
	seatLockTime time.Duration
}

// NewService creates a new railway service
func NewService(repo *Repository, redisClient *redis.Client) *Service {
	return &Service{
		repo:         repo,
		redisClient:  redisClient,
		seatLockTime: 5 * time.Minute, // Default 5 minutes lock time
	}
}

// GetStations retrieves a list of stations with optional city filter
func (s *Service) GetStations(ctx context.Context, city string, page, pageSize int) ([]Station, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	return s.repo.GetStations(ctx, city, page, pageSize)
}

// GetStationByID retrieves a station by ID
func (s *Service) GetStationByID(ctx context.Context, stationID string) (*Station, error) {
	id, err := primitive.ObjectIDFromHex(stationID)
	if err != nil {
		return nil, errors.New("invalid station ID")
	}

	return s.repo.GetStationByID(ctx, id)
}

// GetStationByCode retrieves a station by code
func (s *Service) GetStationByCode(ctx context.Context, code string) (*Station, error) {
	if code == "" {
		return nil, errors.New("station code is required")
	}

	return s.repo.GetStationByCode(ctx, code)
}

// SearchTrains searches for trains matching the criteria
func (s *Service) SearchTrains(ctx context.Context, req SearchTrainsRequest) ([]TrainResponse, error) {
	// Validate the search request
	if req.Origin == "" || req.Destination == "" {
		return nil, errors.New("origin and destination are required")
	}

	if req.Origin == req.Destination {
		return nil, errors.New("origin and destination cannot be the same")
	}

	if req.Passengers <= 0 {
		req.Passengers = 1
	}

	// Ensure date is valid
	if req.Date.IsZero() {
		req.Date = time.Now().Add(24 * time.Hour) // Default to tomorrow
	}

	// Perform direct search first
	trains, err := s.repo.SearchTrains(ctx, req)
	if err != nil {
		return nil, err
	}

	// If no direct trains, try searching for trains that pass through both stations
	if len(trains) == 0 {
		trains, err = s.repo.SearchTrainsByStations(ctx, req.Origin, req.Destination, req.Date)
		if err != nil {
			return nil, err
		}
	}

	// Format response
	var response []TrainResponse
	for _, train := range trains {
		// Format the train data
		trainResp := TrainResponse{
			Train:         &train,
			Duration:      formatDuration(train.Duration),
			DepartureStr:  train.DepartureTime.Format("15:04"),
			ArrivalStr:    train.ArrivalTime.Format("15:04"),
			DepartureDate: train.DepartureTime.Format("Mon, 02 Jan"),
			ArrivalDate:   train.ArrivalTime.Format("Mon, 02 Jan"),
		}

		response = append(response, trainResp)
	}

	return response, nil
}

// GetTrainByID gets detailed information about a train
func (s *Service) GetTrainByID(ctx context.Context, trainID string) (*TrainResponse, error) {
	id, err := primitive.ObjectIDFromHex(trainID)
	if err != nil {
		return nil, errors.New("invalid train ID")
	}

	train, err := s.repo.GetTrainByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Format the train data
	response := &TrainResponse{
		Train:         train,
		Duration:      formatDuration(train.Duration),
		DepartureStr:  train.DepartureTime.Format("15:04"),
		ArrivalStr:    train.ArrivalTime.Format("15:04"),
		DepartureDate: train.DepartureTime.Format("Mon, 02 Jan"),
		ArrivalDate:   train.ArrivalTime.Format("Mon, 02 Jan"),
	}

	return response, nil
}

// GetTrainSeats retrieves the seats for a train
func (s *Service) GetTrainSeats(ctx context.Context, trainID string, class string) ([]TrainSeat, error) {
	id, err := primitive.ObjectIDFromHex(trainID)
	if err != nil {
		return nil, errors.New("invalid train ID")
	}

	return s.repo.GetTrainSeats(ctx, id, class)
}

// GetTrainStops retrieves the stops for a train
func (s *Service) GetTrainStops(ctx context.Context, trainID string) ([]TrainStop, error) {
	id, err := primitive.ObjectIDFromHex(trainID)
	if err != nil {
		return nil, errors.New("invalid train ID")
	}

	return s.repo.GetTrainStops(ctx, id)
}

// LockTrainSeats temporarily locks seats for a booking
func (s *Service) LockTrainSeats(ctx context.Context, trainID string, seatIDs []string, userID string) error {
	if trainID == "" {
		return errors.New("train ID is required")
	}

	if len(seatIDs) == 0 {
		return errors.New("at least one seat must be selected")
	}

	if userID == "" {
		return errors.New("user ID is required")
	}

	// Convert string IDs to ObjectIDs
	_, err := primitive.ObjectIDFromHex(trainID)
	if err != nil {
		return errors.New("invalid train ID")
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
		lockKey := fmt.Sprintf("train_seat_lock:%s:%s", trainID, seatID)

		// Check if the lock exists
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val != "" && val != userID {
			// Seat is locked by someone else
			return errors.New("one or more selected seats are no longer available")
		}
	}

	// Set Redis locks
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("train_seat_lock:%s:%s", trainID, seatID)

		// Set the lock with expiration
		err := s.redisClient.Set(ctx, lockKey, userID, s.seatLockTime).Err()
		if err != nil {
			// If any lock fails, clean up and return error
			s.unlockSeats(ctx, trainID, seatIDs, userID)
			return errors.New("failed to lock seats: " + err.Error())
		}
	}

	return nil
}

// unlockSeats helper to remove Redis locks
func (s *Service) unlockSeats(ctx context.Context, trainID string, seatIDs []string, userID string) {
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("train_seat_lock:%s:%s", trainID, seatID)

		// Only remove if the lock belongs to this user
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val == userID {
			s.redisClient.Del(ctx, lockKey)
		}
	}
}

// ConfirmSeats permanently reserves seats after payment
func (s *Service) ConfirmSeats(ctx context.Context, trainID string, seatIDs []string) error {
	// Convert string IDs to ObjectIDs
	trainObjID, err := primitive.ObjectIDFromHex(trainID)
	if err != nil {
		return errors.New("invalid train ID")
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
		lockKey := fmt.Sprintf("train_seat_lock:%s:%s", trainID, seatID)
		s.redisClient.Del(ctx, lockKey)
	}

	// Update the database
	return s.repo.LockTrainSeats(ctx, trainObjID, seatObjIDs)
}

// ReleaseSeats releases previously locked seats
func (s *Service) ReleaseSeats(ctx context.Context, trainID string, seatIDs []string, userID string) error {
	// Convert string IDs to ObjectIDs
	trainObjID, err := primitive.ObjectIDFromHex(trainID)
	if err != nil {
		return errors.New("invalid train ID")
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
		lockKey := fmt.Sprintf("train_seat_lock:%s:%s", trainID, seatID)

		// Only remove if the lock belongs to this user
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val == userID {
			s.redisClient.Del(ctx, lockKey)
		}
	}

	// Update the database if these were permanent locks
	return s.repo.UnlockTrainSeats(ctx, trainObjID, seatObjIDs)
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
