package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ansh0014/booking/Platform/event"
	"github.com/ansh0014/booking/utils"
	"github.com/gorilla/mux"
)

// SearchEventsHandler handles searching for events
func SearchEventsHandler(w http.ResponseWriter, r *http.Request) {
	var req event.SearchEventsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get event service from context
	eventService := r.Context().Value("eventService").(*event.Service)

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	events, total, err := eventService.SearchEvents(r.Context(), req, page, pageSize)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to search events: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"events": events,
			"count":  len(events),
			"total":  total,
			"page":   page,
		},
	})
}

// GetEventDetailsHandler retrieves details for a specific event
func GetEventDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]

	eventService := r.Context().Value("eventService").(*event.Service)

	eventDetails, err := eventService.GetEventByID(r.Context(), eventID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Event not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    eventDetails,
	})
}

// GetEventSeatsHandler retrieves available seats for an event
func GetEventSeatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]

	// Optional ticket type filter
	ticketTypeID := r.URL.Query().Get("ticket_type_id")

	eventService := r.Context().Value("eventService").(*event.Service)

	seats, err := eventService.GetEventSeats(r.Context(), eventID, ticketTypeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get event seats: "+err.Error())
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

// GetEventTicketTypesHandler retrieves ticket types for an event
func GetEventTicketTypesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]

	eventService := r.Context().Value("eventService").(*event.Service)

	ticketTypes, err := eventService.GetTicketTypes(r.Context(), eventID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get ticket types: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"ticket_types": ticketTypes,
		},
	})
}

// LockEventSeatsHandler temporarily reserves seats for an event
func LockEventSeatsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EventID      string   `json:"event_id"`
		TicketTypeID string   `json:"ticket_type_id"`
		SeatIDs      []string `json:"seat_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get the user ID from the authenticated request
	userID := r.Context().Value("userID").(string)

	eventService := r.Context().Value("eventService").(*event.Service)

	err := eventService.LockEventSeats(r.Context(), req.EventID, req.TicketTypeID, req.SeatIDs, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to lock seats: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Seats locked successfully",
		"data": map[string]interface{}{
			"event_id":       req.EventID,
			"ticket_type_id": req.TicketTypeID,
			"seat_ids":       req.SeatIDs,
		},
	})
}
