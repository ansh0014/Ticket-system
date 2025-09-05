package config

import (
    "errors"
    "os"
    "strconv"
    "time"
)

type Config struct {
    AuthURL      string
    BookingURL   string
    PaymentURL   string
    VenueURL     string
    JWTSecret    string
    Port         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}

// Load reads configuration from environment variables and returns a Config.
func Load() (*Config, error) {
    c := &Config{
        AuthURL:    os.Getenv("AUTH_SERVICE_URL"),
        BookingURL: os.Getenv("BOOKING_SERVICE_URL"),
        PaymentURL: os.Getenv("PAYMENT_SERVICE_URL"),
        VenueURL:   os.Getenv("VENUE_SERVICE_URL"),
        JWTSecret:  os.Getenv("JWT_SECRET"),
        Port:       os.Getenv("GATEWAY_PORT"),
    }

    if c.Port == "" {
        c.Port = "8080"
    }

    // timeouts (in seconds)
    c.ReadTimeout = parseEnvDuration("GATEWAY_READ_TIMEOUT", 15*time.Second)
    c.WriteTimeout = parseEnvDuration("GATEWAY_WRITE_TIMEOUT", 20*time.Second)
    c.IdleTimeout = parseEnvDuration("GATEWAY_IDLE_TIMEOUT", 60*time.Second)

    if err := c.Validate(); err != nil {
        return nil, err
    }
    return c, nil
}

func parseEnvDuration(name string, def time.Duration) time.Duration {
    if v := os.Getenv(name); v != "" {
        if s, err := strconv.Atoi(v); err == nil && s >= 0 {
            return time.Duration(s) * time.Second
        }
    }
    return def
}

// Validate ensures required fields are provided.
func (c *Config) Validate() error {
    if c.AuthURL == "" {
        return errors.New("AUTH_SERVICE_URL is required")
    }
    if c.BookingURL == "" {
        return errors.New("BOOKING_SERVICE_URL is required")
    }
    if c.PaymentURL == "" {
        return errors.New("PAYMENT_SERVICE_URL is required")
    }
    if c.VenueURL == "" {
        return errors.New("VENUE_SERVICE_URL is required")
    }
    return nil
}
