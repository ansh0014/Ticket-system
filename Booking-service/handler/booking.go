package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ansh0014/booking/model"
	"github.com/ansh0014/booking/service"
	"github.com/ansh0014/booking/utils"
)

// CreateBookingHandler creates a new booking
func CreateBookingHandler(w http.ResponseWriter, r *http.Request) {
	var req model.BookingRequest

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
	userID, err := utils.GetUserFromContext(r.Context())
	if err != nil {
		utils.UnauthorizedResponse(w, "User not authenticated")
		return
	}

	// Get booking service
	bookingService := r.Context().Value("bookingService").(*service.BookingService)

	// Create booking
	booking, err := bookingService.CreateBooking(r.Context(), req, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return the created booking
	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Booking created successfully",
		"data": map[string]interface{}{
			"booking": booking,
		},
	})
}

// GetBookingHandler retrieves a booking by ID
func GetBookingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookingID := vars["id"]

	// Get user ID from context
	userID, err := utils.GetUserFromContext(r.Context())
	if err != nil {
		utils.UnauthorizedResponse(w, "User not authenticated")
		return
	}

	// Get booking service
	bookingService := r.Context().Value("bookingService").(*service.BookingService)

	// Get booking
	booking, err := bookingService.GetBooking(r.Context(), bookingID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Booking not found")
		return
	}

	// Check if the booking belongs to the user
	if booking.UserID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "You don't have permission to access this booking")
		return
	}

	// Return the booking
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"booking": booking,
		},
	})
}

// GetUserBookingsHandler retrieves all bookings for a user
func GetUserBookingsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := utils.GetUserFromContext(r.Context())
	if err != nil {
		utils.UnauthorizedResponse(w, "User not authenticated")
		return
	}

	// Parse query parameters
	page, pageSize := utils.GetPageParams(r)

	// Get booking service
	bookingService := r.Context().Value("bookingService").(*service.BookingService)

	// Get user's bookings
	bookings, total, err := bookingService.GetUserBookings(r.Context(), userID, page, pageSize)
	if err != nil {
		utils.ServerErrorResponse(w, "Failed to get bookings: "+err.Error())
		return
	}

	// Return the bookings
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"bookings": bookings,
			"count":    len(bookings),
			"total":    total,
			"page":     page,
		},
	})
}

// CancelBookingHandler cancels a booking
func CancelBookingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookingID := vars["id"]

	// Get user ID from context
	userID, err := utils.GetUserFromContext(r.Context())
	if err != nil {
		utils.UnauthorizedResponse(w, "User not authenticated")
		return
	}

	// Get booking service
	bookingService := r.Context().Value("bookingService").(*service.BookingService)

	// Get booking first to check ownership
	booking, err := bookingService.GetBooking(r.Context(), bookingID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Booking not found")
		return
	}

	// Check if the booking belongs to the user
	if booking.UserID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "You don't have permission to cancel this booking")
		return
	}

	// Cancel the booking
	err = bookingService.CancelBooking(r.Context(), bookingID)
	if err != nil {
		utils.ServerErrorResponse(w, "Failed to cancel booking: "+err.Error())
		return
	}

	// Return success
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Booking cancelled successfully",
	})
}
