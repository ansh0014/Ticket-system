package routes

import (
    "net/http"

    "github.com/gorilla/mux"

    "github.com/ansh0014/venue/handler"
)

func NewRouter(h *handler.Handler) http.Handler {
    r := mux.NewRouter()

    // Venue
    r.HandleFunc("/venues", h.CreateVenue).Methods("POST")
    r.HandleFunc("/venues", h.ListVenues).Methods("GET")
    r.HandleFunc("/venues/{id}", h.GetVenue).Methods("GET")

    // Halls
    r.HandleFunc("/venues/{id}/halls", h.CreateHall).Methods("POST")
    r.HandleFunc("/venues/{id}/halls", h.ListHalls).Methods("GET")

    // Seats
    r.HandleFunc("/halls/{id}/seats", h.AddSeat).Methods("POST")
    r.HandleFunc("/halls/{id}/seats", h.ListSeats).Methods("GET")

    // health
    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    }).Methods("GET")

    return r
}