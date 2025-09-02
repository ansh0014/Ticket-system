package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ansh0014/booking/Platform/flight"
	"github.com/ansh0014/booking/utils"
	"github.com/gorilla/mux"
)

// SearchFlightsHandler handles searching for flights
func SearchFlightsHandler(w http.ResponseWriter, r *http.Request) {
	var req flight.SearchFlightsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequestResponse(w, "Invalid request body", nil)
		return
	}

	// Get flight service from context
	flightService := r.Context().Value("flightService").(*flight.Service)

	flights, err := flightService.SearchFlights(r.Context(), req)
	if err != nil {
		utils.ServerErrorResponse(w, "Failed to search flights: "+err.Error())
		return
	}

	utils.OkResponse(w, "Success", map[string]interface{}{
		"flights": flights,
		"count":   len(flights),
	})
}

// GetFlightDetailsHandler retrieves details for a specific flight
func GetFlightDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flightID := vars["id"]

	flightService := r.Context().Value("flightService").(*flight.Service)

	flightDetails, err := flightService.GetFlightByID(r.Context(), flightID)
	if err != nil {
		utils.NotFoundResponse(w, "Flight not found")
		return
	}

	utils.OkResponse(w, "Success", flightDetails)
}

// GetFlightSeatsHandler retrieves available seats for a flight
func GetFlightSeatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flightID := vars["id"]

	flightService := r.Context().Value("flightService").(*flight.Service)

	seats, err := flightService.GetFlightSeats(r.Context(), flightID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get flight seats: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"seats": seats,
			"count": len(seats),
		},
	})
}

// LockFlightSeatsHandler temporarily reserves seats for a flight
func LockFlightSeatsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FlightID string   `json:"flight_id"`
		SeatIDs  []string `json:"seat_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequestResponse(w, "Invalid request body", nil)
		return
	}

	// Get the user ID from the authenticated request
	userID := r.Context().Value("userID").(string)

	flightService := r.Context().Value("flightService").(*flight.Service)

	err := flightService.LockFlightSeats(r.Context(), req.FlightID, req.SeatIDs, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to lock seats: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Seats locked successfully",
		"data": map[string]interface{}{
			"flight_id": req.FlightID,
			"seat_ids":  req.SeatIDs,
		},
	})
}
