package router

import (
    "net/http"

    "github.com/ansh0014/booking/handler"
    "github.com/gorilla/mux"
)

func SetupRoutes() http.Handler {
    r := mux.NewRouter()
    
    // Booking endpoint
    r.HandleFunc("/api/bookings", handler.CreateBookingHandler).Methods("POST")
    r.HandleFunc("/api/bookings/{id}", handler.GetBookingHandler).Methods("GET")
    r.HandleFunc("/api/bookings/{id}/cancel", handler.CancelBookingHandler).Methods("POST")
    r.HandleFunc("/api/bookings/{id}/confirm", handler.ConfirmBookingHandler).Methods("POST")
    r.HandleFunc("/api/users/{userId}/bookings", handler.GetUserBookingsHandler).Methods("GET")
    
    // Seat endpoint
    r.HandleFunc("/api/seats/lock", handler.LockSeatsHandler).Methods("POST")
    r.HandleFunc("/api/shows/availability", handler.GetAvailabilityHandler).Methods("POST")
    
    
    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }).Methods("GET")
    
    return r
}