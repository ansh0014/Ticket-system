package handler

import (
    "context"
    "net/http"
    "strconv"

    "github.com/ansh0014/venue/model"
    "github.com/ansh0014/venue/service"
    "github.com/ansh0014/venue/utils"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
    svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
    return &Handler{svc: svc}
}

// CreateVenue POST /venues
func (h *Handler) CreateVenue(w http.ResponseWriter, r *http.Request) {
    var v model.Venue
    if err := utils.ReadJSON(r, &v); err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "invalid request: "+err.Error())
        return
    }
    created, err := h.svc.CreateVenue(r.Context(), &v)
    if err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }
    utils.RespondWithJSON(w, http.StatusCreated, created)
}

// ListVenues GET /venues?page=1&page_size=20
func (h *Handler) ListVenues(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    page, _ := strconv.Atoi(q.Get("page"))
    pageSize, _ := strconv.Atoi(q.Get("page_size"))
    filter := map[string]interface{}{}
    if city := q.Get("city"); city != "" {
        filter["city"] = city
    }
    venues, total, err := h.svc.ListVenues(context.Background(), filter, page, pageSize)
    if err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }
    utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
        "data":  venues,
        "total": total,
        "page":  page,
    })
}

// GetVenue GET /venues/{id}
func (h *Handler) GetVenue(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]
    id, err := primitive.ObjectIDFromHex(idStr)
    if err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "invalid id")
        return
    }
    v, err := h.svc.GetVenue(r.Context(), id)
    if err != nil {
        utils.RespondWithError(w, http.StatusNotFound, "venue not found")
        return
    }
    utils.RespondWithJSON(w, http.StatusOK, v)
}

// CreateHall POST /venues/{id}/halls
func (h *Handler) CreateHall(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    venueIDStr := vars["id"]
    venueID, err := primitive.ObjectIDFromHex(venueIDStr)
    if err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "invalid venue id")
        return
    }
    var hall model.Hall
    if err := utils.ReadJSON(r, &hall); err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "invalid request: "+err.Error())
        return
    }
    hall.VenueID = venueID
    created, err := h.svc.CreateHall(r.Context(), &hall)
    if err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }
    utils.RespondWithJSON(w, http.StatusCreated, created)
}

// ListHalls GET /venues/{id}/halls
func (h *Handler) ListHalls(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    venueIDStr := vars["id"]
    venueID, err := primitive.ObjectIDFromHex(venueIDStr)
    if err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "invalid venue id")
        return
    }
    halls, err := h.svc.ListHalls(r.Context(), venueID)
    if err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }
    utils.RespondWithJSON(w, http.StatusOK, halls)
}

// AddSeat POST /halls/{id}/seats
func (h *Handler) AddSeat(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    hallIDStr := vars["id"]
    hallID, err := primitive.ObjectIDFromHex(hallIDStr)
    if err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "invalid hall id")
        return
    }
    var seat model.Seat
    if err := utils.ReadJSON(r, &seat); err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "invalid request: "+err.Error())
        return
    }
    seat.HallID = hallID
    created, err := h.svc.AddSeat(r.Context(), &seat)
    if err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }
    utils.RespondWithJSON(w, http.StatusCreated, created)
}

// ListSeats GET /halls/{id}/seats
func (h *Handler) ListSeats(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    hallIDStr := vars["id"]
    hallID, err := primitive.ObjectIDFromHex(hallIDStr)
    if err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "invalid hall id")
        return
    }
    seats, err := h.svc.ListSeats(r.Context(), hallID)
    if err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }
    utils.RespondWithJSON(w, http.StatusOK, seats)
}