package router

import (
	"context"
	"net/http"

	"github.com/ansh0014/booking/handler"
	// "github.com/ansh0014/booking/utils"
	"github.com/gorilla/mux"
)

// SetupRoutes configures all routes for the booking service
func SetupRoutes(platformServices map[string]interface{}) http.Handler {
	r := mux.NewRouter()

	// Middleware to inject platform services
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Add all platform services to the request context
			ctx := req.Context()
			for name, service := range platformServices {
				ctx = context.WithValue(ctx, name+"Service", service)
			}
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})

	// // Apply global middlewares
	// r.Use(utils.LoggingMiddleware)
	// r.Use(utils.RecoverMiddleware)
	// r.Use(utils.CORSMiddleware)

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Flight routes
	r.HandleFunc("/api/platforms/flight/search", handler.SearchFlightsHandler).Methods("POST")
	r.HandleFunc("/api/platforms/flight/{id}", handler.GetFlightDetailsHandler).Methods("GET")
	r.HandleFunc("/api/platforms/flight/{id}/seats", handler.GetFlightSeatsHandler).Methods("GET")
	r.HandleFunc("/api/platforms/flight/seats/lock", handler.LockFlightSeatsHandler).Methods("POST")

	// Railway routes
	r.HandleFunc("/api/platforms/railway/search", handler.SearchTrainsHandler).Methods("POST")
	r.HandleFunc("/api/platforms/railway/{id}", handler.GetTrainDetailsHandler).Methods("GET")
	r.HandleFunc("/api/platforms/railway/{id}/seats", handler.GetTrainSeatsHandler).Methods("GET")
	r.HandleFunc("/api/platforms/railway/{id}/stops", handler.GetTrainStopsHandler).Methods("GET")
	r.HandleFunc("/api/platforms/railway/seats/lock", handler.LockTrainSeatsHandler).Methods("POST")

	// Event routes
	r.HandleFunc("/api/platforms/event/search", handler.SearchEventsHandler).Methods("POST")
	r.HandleFunc("/api/platforms/event/{id}", handler.GetEventDetailsHandler).Methods("GET")
	r.HandleFunc("/api/platforms/event/{id}/seats", handler.GetEventSeatsHandler).Methods("GET")
	r.HandleFunc("/api/platforms/event/{id}/ticket-types", handler.GetEventTicketTypesHandler).Methods("GET")
	r.HandleFunc("/api/platforms/event/seats/lock", handler.LockEventSeatsHandler).Methods("POST")

	// Movie routes
	r.HandleFunc("/api/platforms/movie", handler.GetMoviesHandler).Methods("GET")
	r.HandleFunc("/api/platforms/movie/search", handler.SearchMoviesHandler).Methods("POST")
	r.HandleFunc("/api/platforms/movie/{id}", handler.GetMovieDetailsHandler).Methods("GET")
	r.HandleFunc("/api/platforms/movie/{id}/shows", handler.GetMovieShowsHandler).Methods("GET")
	r.HandleFunc("/api/platforms/movie/seats", handler.GetMovieSeatsHandler).Methods("GET")
	r.HandleFunc("/api/platforms/movie/seats/lock", handler.LockMovieSeatsHandler).Methods("POST")

	// Generic seat locking (works for all platforms)
	r.HandleFunc("/api/seats/lock", handler.LockSeatsHandler).Methods("POST")

	// Booking routes (works for all platforms)
	r.HandleFunc("/api/bookings", handler.CreateBookingHandler).Methods("POST")
	r.HandleFunc("/api/bookings/{id}", handler.GetBookingHandler).Methods("GET")
	r.HandleFunc("/api/bookings/{id}/cancel", handler.CancelBookingHandler).Methods("POST")
	r.HandleFunc("/api/users/me/bookings", handler.GetUserBookingsHandler).Methods("GET")

	return r
}
