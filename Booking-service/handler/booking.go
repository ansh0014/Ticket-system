package handler

import (
    "encoding/json"
    "net/http"

    "github.com/ansh0014/booking/model"
    "github.com/ansh0014/booking/service"
    "github.com/gorilla/mux"
)

// CreateBookingHandler handles booking creation
func CreateBookingHandler(w http.ResponseWriter, r *http.Request) {
    var req model.CreateBookingRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    booking, err := service.CreateBooking(&req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    
    expiresIn := int(booking.ExpiryTime.Sub(booking.BookingTime).Seconds())
    
    // TODO payment URL (call Payment gateway)
    paymentURL := "http://payment-service/pay/" + booking.ID
    
    response := model.BookingResponse{
        Booking:    *booking,
        PaymentURL: paymentURL,
        ExpiresIn:  expiresIn,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// GetBookingHandler retrieves booking details
func GetBookingHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    bookingID := vars["id"]
    
    booking, err := service.GetBooking(bookingID)
    if err != nil {
        http.Error(w, "Booking not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(booking)
}

// CancelBookingHandler cancels a booking
func CancelBookingHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    bookingID := vars["id"]
    
    err := service.CancelBooking(bookingID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Booking cancelled successfully",
    })
}

// GetUserBookingsHandler gets all bookings for a user
func GetUserBookingsHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    userID := vars["userId"]
    
    bookings, err := service.GetUserBookings(userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bookings)
}

// ConfirmBookingHandler confirms a booking after payment
func ConfirmBookingHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    bookingID := vars["id"]
    
    var req struct {
        PaymentID string `json:"payment_id"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    err := service.ConfirmBooking(bookingID, req.PaymentID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Booking confirmed successfully",
    })
}