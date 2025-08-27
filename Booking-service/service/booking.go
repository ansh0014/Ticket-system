package service

import (
    "context"
    "errors"
    "time"

    "github.com/ansh0014/booking/config"
    "github.com/ansh0014/booking/model"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateBooking creates a new booking with initial status "pending"
func CreateBooking(req *model.CreateBookingRequest) (*model.Booking, error) {
    // First lock the seats
    err := LockSeats(req.ShowID, req.Seats, req.UserID, DefaultLockDuration)
    if err != nil {
        return nil, err
    }

    // TODO: Get price information from Venue service (for now use fixed price)
    seatPrice := 100.0 // Fixed price per seat for now
    totalPrice := seatPrice * float64(len(req.Seats))

    // Create booking
    booking := &model.Booking{
        ID:          primitive.NewObjectID().Hex(),
        UserID:      req.UserID,
        ShowID:      req.ShowID,
        Seats:       req.Seats,
        TotalPrice:  totalPrice,
        Status:      "pending",
        BookingTime: time.Now(),
        ExpiryTime:  time.Now().Add(time.Duration(DefaultLockDuration) * time.Second),
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    // Insert into MongoDB
    _, err = config.MongoDB.Collection("bookings").InsertOne(context.Background(), booking)
    if err != nil {
        // If insert fails, unlock the seats
        UnlockSeats(req.ShowID, req.Seats, req.UserID)
        return nil, err
    }

    return booking, nil
}

// GetBooking retrieves a booking by ID
func GetBooking(bookingID string) (*model.Booking, error) {
    var booking model.Booking
    err := config.MongoDB.Collection("bookings").FindOne(
        context.Background(),
        bson.M{"_id": bookingID},
    ).Decode(&booking)
    
    if err != nil {
        return nil, err
    }
    return &booking, nil
}

// UpdateBookingStatus updates the status of a booking
func UpdateBookingStatus(bookingID, status string) error {
    _, err := config.MongoDB.Collection("bookings").UpdateOne(
        context.Background(),
        bson.M{"_id": bookingID},
        bson.M{
            "$set": bson.M{
                "status":     status,
                "updated_at": time.Now(),
            },
        },
    )
    return err
}

// ConfirmBooking marks a booking as confirmed after payment
func ConfirmBooking(bookingID, paymentID string) error {
    booking, err := GetBooking(bookingID)
    if err != nil {
        return err
    }
    
    if booking.Status != "pending" {
        return errors.New("booking is not in pending status")
    }
    
    _, err = config.MongoDB.Collection("bookings").UpdateOne(
        context.Background(),
        bson.M{"_id": bookingID},
        bson.M{
            "$set": bson.M{
                "status":     "confirmed",
                "payment_id": paymentID,
                "updated_at": time.Now(),
            },
        },
    )
    
    if err != nil {
        return err
    }
    
    // TODO: Publish event that booking is confirmed (for other services)
    
    return nil
}

// CancelBooking cancels a booking and unlocks seats
func CancelBooking(bookingID string) error {
    booking, err := GetBooking(bookingID)
    if err != nil {
        return err
    }
    
    if booking.Status == "cancelled" {
        return errors.New("booking is already cancelled")
    }
    
    _, err = config.MongoDB.Collection("bookings").UpdateOne(
        context.Background(),
        bson.M{"_id": bookingID},
        bson.M{
            "$set": bson.M{
                "status":     "cancelled",
                "updated_at": time.Now(),
            },
        },
    )
    
    if err != nil {
        return err
    }
    
    // Unlock seats if the booking was pending
    if booking.Status == "pending" {
        UnlockSeats(booking.ShowID, booking.Seats, booking.UserID)
    }
    
    // TODO: Publish event that booking is cancelled (for other services)
    
    return nil
}

// GetUserBookings gets all bookings for a user
func GetUserBookings(userID string) ([]model.Booking, error) {
    cursor, err := config.MongoDB.Collection("bookings").Find(
        context.Background(),
        bson.M{"user_id": userID},
    )
    
    if err != nil {
        return nil, err
    }
    
    var bookings []model.Booking
    err = cursor.All(context.Background(), &bookings)
    if err != nil {
        return nil, err
    }
    
    return bookings, nil
}