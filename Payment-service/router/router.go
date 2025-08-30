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

    // API documentation
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        w.Write([]byte(`
        <html>
            <head>
                <title>Payment Service API</title>
                <style>
                    body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
                    h1 { color: #333; }
                    h2 { color: #666; margin-top: 30px; }
                    code { background: #f4f4f4; padding: 2px 5px; border-radius: 3px; }
                </style>
            </head>
            <body>
                <h1>Payment Service API</h1>
                <p>Available endpoints:</p>

                <h2>Create Payment</h2>
                <p><code>POST /api/payments</code></p>

                <h2>Get Payment Details</h2>
                <p><code>GET /api/payments/{id}</code></p>

                <h2>Refund Payment</h2>
                <p><code>POST /api/payments/refund</code></p>

                <h2>Verify Payment</h2>
                <p><code>POST /api/payments/verify</code></p>

                <h2>Payment Gateway Webhook</h2>
                <p><code>POST /api/webhook</code></p>
                <p><code>POST /api/webhook/{provider}</code></p>

                <h2>Health Check</h2>
                <p><code>GET /health</code></p>
            </body>
        </html>
        `))
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

        // Log request (in a production environment, use a proper logging library)
        // log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)

        // Call the original handler
        handler.ServeHTTP(w, r)
    })
}