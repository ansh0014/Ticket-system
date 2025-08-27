package handler

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/ansh0014/booking/model"
    "github.com/ansh0014/booking/service"
)

// LockSeatsHandler locks seats temporarily for a user
func LockSeatsHandler(w http.ResponseWriter, r *http.Request) {
    var req model.LockSeatsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    err := service.LockSeats(req.ShowID, req.Seats, req.UserID, req.Duration)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Seats locked successfully",
    })
}

// GetAvailabilityHandler checks availability for a show
func GetAvailabilityHandler(w http.ResponseWriter, r *http.Request) {
    var req model.AvailabilityRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    seatStatus, err := service.GetAvailability(req.ShowID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Organize seats by status
    available := []string{}
    locked := []string{}
    booked := []string{}
    
    // TODO: Get full seat list from Venue service
    // For now, we'll just categorize the ones in Redis
    
    for seatID, userID := range seatStatus {
        if userID != "" {
            locked = append(locked, seatID)
        }
    }
    
    response := model.AvailabilityResponse{
        ShowID:    req.ShowID,
        Available: available,
        Locked:    locked,
        Booked:    booked,
        UpdatedAt: time.Now(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}