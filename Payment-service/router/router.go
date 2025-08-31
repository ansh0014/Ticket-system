package router

import (
	"net/http"

	"github.com/ansh0014/payment/handler"
	"github.com/gorilla/mux"
)

// SetupRoutes configures all API routes for the payment service
func SetupRoutes() http.Handler {
	r := mux.NewRouter()

	// Payment endpoints
	r.HandleFunc("/api/payments", handler.CreatePaymentHandler).Methods("POST")
	r.HandleFunc("/api/payments/{id}", handler.GetPaymentHandler).Methods("GET")
	r.HandleFunc("/api/payments/refund", handler.RefundPaymentHandler).Methods("POST")
	r.HandleFunc("/api/payments/verify", handler.VerifyPaymentHandler).Methods("POST")

	// Webhook endpoints for payment gateway callbacks
	r.HandleFunc("/api/webhook", handler.WebhookHandler).Methods("POST")
	r.HandleFunc("/api/webhook/{provider}", handler.WebhookHandler).Methods("POST")

	// Health check for container orchestration and monitoring
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Simple root endpoint
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Payment Service"))
	}).Methods("GET")

	// CORS middleware
	return addMiddleware(r)
}

// addMiddleware adds CORS and other middleware to the router
func addMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the original handler
		handler.ServeHTTP(w, r)
	})
}
