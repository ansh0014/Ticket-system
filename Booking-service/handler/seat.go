package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ansh0014/booking/model"
	"github.com/ansh0014/booking/service"
	"github.com/ansh0014/booking/utils"
)

// LockSeatsHandler is a generic handler that works with any platform
func LockSeatsHandler(w http.ResponseWriter, r *http.Request) {
	var req model.SeatLockRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate platform
	if req.Platform == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Platform is required")
		return
	}

	// Validate platform ID
	if req.PlatformID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Platform ID is required")
		return
	}

	// Validate seats
	if len(req.SeatIDs) == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "At least one seat must be selected")
		return
	}

	// Get user ID from context
	userID := r.Context().Value("userID").(string)

	// Get seat service
	seatService := r.Context().Value("seatService").(*service.SeatService)

	// Lock seats
	err := seatService.LockSeats(r.Context(), req, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return success
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Seats locked successfully",
		"data": map[string]interface{}{
			"platform":    req.Platform,
			"platform_id": req.PlatformID,
			"seat_ids":    req.SeatIDs,
			"expires_at":  time.Now().Add(5 * time.Minute),
		},
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
