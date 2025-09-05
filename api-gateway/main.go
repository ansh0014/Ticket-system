package main

import (
    "log"
    "net/http"
    "os"

    "github.com/ansh0014/api/middleware"
    "github.com/ansh0014/api/pkg"
    "github.com/ansh0014/api/routes"

    "github.com/gorilla/handlers"
)

func main() {
    authURL := os.Getenv("AUTH_SERVICE_URL")
    bookingURL := os.Getenv("BOOKING_SERVICE_URL")
    paymentURL := os.Getenv("PAYMENT_SERVICE_URL")
    venueURL := os.Getenv("VENUE_SERVICE_URL")
    port := os.Getenv("GATEWAY_PORT")
    if port == "" {
        port = "8080"
    }

    if authURL == "" || bookingURL == "" || paymentURL == "" || venueURL == "" {
        log.Fatal("set AUTH_SERVICE_URL, BOOKING_SERVICE_URL, PAYMENT_SERVICE_URL and VENUE_SERVICE_URL")
    }

    pm := pkg.NewProxyMap(map[string]string{
        "/auth/":    authURL,
        "/booking/": bookingURL,
        "/payment/": paymentURL,
        "/venue/":   venueURL,
    })

    r := routes.NewRouter(pm)

    // apply middlewares: CORS + JWT extract + rate limit + logging
    cors := handlers.CORS(
        handlers.AllowedOrigins([]string{"*"}),
        handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
        handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-Requested-With", "X-User-ID"}),
    )

    h := cors(middleware.JWTExtract(r))
    h = middleware.RateLimit(h)

    log.Printf("api-gateway listening on :%s", port)
    logged := handlers.CombinedLoggingHandler(log.Writer(), h)
    if err := http.ListenAndServe(":"+port, logged); err != nil {
        log.Fatal(err)
    }
}