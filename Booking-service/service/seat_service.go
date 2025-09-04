package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ansh0014/booking/model"
	"github.com/go-redis/redis/v8"
)

// SeatService provides seat-related functionality
type SeatService struct {
	redisClient      *redis.Client
	platformServices map[string]interface{}
}

// NewSeatService creates a new seat service
func NewSeatService(redisClient *redis.Client, platformServices map[string]interface{}) *SeatService {
	return &SeatService{
		redisClient:      redisClient,
		platformServices: platformServices,
	}
}

// LockSeats locks seats for a specific platform and ID
func (s *SeatService) LockSeats(ctx context.Context, req model.SeatLockRequest, userID string) error {
	if req.Platform == "" {
		return errors.New("platform is required")
	}

	if req.PlatformID == "" {
		return errors.New("platform ID is required")
	}

	if len(req.SeatIDs) == 0 {
		return errors.New("at least one seat must be selected")
	}

	// Create a key format based on platform and ID
	keyPrefix := fmt.Sprintf("%s:%s", req.Platform, req.PlatformID)

	// Check if seats are already locked or booked
	for _, seatID := range req.SeatIDs {
		key := fmt.Sprintf("%s:seat:%s", keyPrefix, seatID)
		val, err := s.redisClient.Get(ctx, key).Result()
		if err == nil && val != "" && val != userID {
			return errors.New("one or more seats are already locked")
		}
	}

	// Lock all seats
	for _, seatID := range req.SeatIDs {
		key := fmt.Sprintf("%s:seat:%s", keyPrefix, seatID)
		err := s.redisClient.Set(ctx, key, userID, 5*time.Minute).Err()
		if err != nil {
			// Clean up on error
			s.ReleaseSeats(ctx, req, userID)
			return err
		}
	}

	return nil
}

// ReleaseSeats unlocks seats for a specific platform and ID
func (s *SeatService) ReleaseSeats(ctx context.Context, req model.SeatLockRequest, userID string) error {
	if req.Platform == "" {
		return errors.New("platform is required")
	}

	if req.PlatformID == "" {
		return errors.New("platform ID is required")
	}

	if len(req.SeatIDs) == 0 {
		return errors.New("at least one seat must be selected")
	}

	// Create a key format based on platform and ID
	keyPrefix := fmt.Sprintf("%s:%s", req.Platform, req.PlatformID)

	// Release seats
	for _, seatID := range req.SeatIDs {
		key := fmt.Sprintf("%s:seat:%s", keyPrefix, seatID)
		// Only release if locked by this user
		val, err := s.redisClient.Get(ctx, key).Result()
		if err == nil && val == userID {
			s.redisClient.Del(ctx, key)
		}
	}

	return nil
}

// GetAvailability returns all seat statuses for a specific show ID
func (s *SeatService) GetAvailability(ctx context.Context, showID string) (map[string]string, error) {
	if showID == "" {
		return nil, errors.New("show ID is required")
	}

	pattern := fmt.Sprintf("*:%s:seat:*", showID)
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, key := range keys {
		// Extract seat ID from key
		parts := splitKey(key)
		if len(parts) >= 4 {
			seatID := parts[len(parts)-1]
			userID, err := s.redisClient.Get(ctx, key).Result()
			if err == nil {
				result[seatID] = userID
			}
		}
	}

	return result, nil
}

// Helper function to split a Redis key
func splitKey(key string) []string {
	var result []string
	var current string

	for _, r := range key {
		if r == ':' {
			result = append(result, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}

	return result
}
