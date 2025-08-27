package service

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/ansh0014/booking/config"
)

const (
    DefaultLockDuration = 200
)

// LockSeats locks a set of seats for a specific user
func LockSeats(showID string, seats []string, userID string, duration int) error {
    if duration <= 0 {
        duration = DefaultLockDuration
    }

    // Check if seats are already locked or booked
    for _, seat := range seats {
        key := fmt.Sprintf("seat:%s:%s", showID, seat)
        val, err := config.RedisClient.Get(context.Background(), key).Result()
        if err == nil && val != "" && val != userID {
            return errors.New("one or more seats are already locked")
        }
    }

    // Lock all seats
    for _, seat := range seats {
        key := fmt.Sprintf("seat:%s:%s", showID, seat)
        err := config.RedisClient.Set(context.Background(), key, userID, time.Duration(duration)*time.Second).Err()
        if err != nil {
            return err
        }
    }

    return nil
}

// UnlockSeats unlocks seats for a specific user
func UnlockSeats(showID string, seats []string, userID string) error {
    for _, seat := range seats {
        key := fmt.Sprintf("seat:%s:%s", showID, seat)
        
        // Only unlock if the seat is locked by this user
        val, err := config.RedisClient.Get(context.Background(), key).Result()
        if err == nil && val == userID {
            config.RedisClient.Del(context.Background(), key)
        }
    }
    return nil
}

// GetSeatStatus checks if a seat is locked and by whom
func GetSeatStatus(showID string, seat string) (string, bool, error) {
    key := fmt.Sprintf("seat:%s:%s", showID, seat)
    userID, err := config.RedisClient.Get(context.Background(), key).Result()
    if err != nil {
        return "", false, nil // Seat is not locked
    }
    return userID, true, nil // Seat is locked by userID
}

// GetAvailability returns all seat statuses for a show
func GetAvailability(showID string) (map[string]string, error) {
    pattern := fmt.Sprintf("seat:%s:*", showID)
    keys, err := config.RedisClient.Keys(context.Background(), pattern).Result()
    if err != nil {
        return nil, err
    }

    result := make(map[string]string)
    for _, key := range keys {
        // Extract seat ID from key
        seatID := key[len(fmt.Sprintf("seat:%s:", showID)):]
        userID, err := config.RedisClient.Get(context.Background(), key).Result()
        if err == nil {
            result[seatID] = userID
        }
    }
    
    return result, nil
}