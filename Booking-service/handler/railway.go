package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ansh0014/booking/Platform/railway"
	"github.com/ansh0014/booking/utils"
	"github.com/gorilla/mux"
)

// SearchTrainsHandler handles searching for trains
func SearchTrainsHandler(w http.ResponseWriter, r *http.Request) {
	var req railway.SearchTrainsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get railway service from context
	railwayService := r.Context().Value("railwayService").(*railway.Service)

	trains, err := railwayService.SearchTrains(r.Context(), req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to search trains: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"trains": trains,
			"count":  len(trains),
		},
	})
}

// GetTrainDetailsHandler retrieves details for a specific train
func GetTrainDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainID := vars["id"]

	railwayService := r.Context().Value("railwayService").(*railway.Service)

	trainDetails, err := railwayService.GetTrainByID(r.Context(), trainID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Train not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    trainDetails,
	})
}

// GetTrainSeatsHandler retrieves available seats for a train
func GetTrainSeatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainID := vars["id"]

	// Optional class filter
	class := r.URL.Query().Get("class")

	railwayService := r.Context().Value("railwayService").(*railway.Service)

	seats, err := railwayService.GetTrainSeats(r.Context(), trainID, class)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get train seats: "+err.Error())
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

// GetTrainStopsHandler retrieves all stops for a train
func GetTrainStopsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainID := vars["id"]

	railwayService := r.Context().Value("railwayService").(*railway.Service)

	stops, err := railwayService.GetTrainStops(r.Context(), trainID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get train stops: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"stops": stops,
			"count": len(stops),
		},
	})
}

// LockTrainSeatsHandler temporarily reserves seats for a train
func LockTrainSeatsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TrainID string   `json:"train_id"`
		SeatIDs []string `json:"seat_ids"`
		Class   string   `json:"class,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get the user ID from the authenticated request
	userID := r.Context().Value("userID").(string)

	railwayService := r.Context().Value("railwayService").(*railway.Service)

	err := railwayService.LockTrainSeats(r.Context(), req.TrainID, req.SeatIDs, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to lock seats: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Seats locked successfully",
		"data": map[string]interface{}{
			"train_id": req.TrainID,
			"seat_ids": req.SeatIDs,
			"class":    req.Class,
		},
	})
}
