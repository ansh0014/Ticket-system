package service

import (
	"context"
	"errors"
	"time"

	"github.com/ansh0014/booking/model"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BookingService handles booking-related functionality
type BookingService struct {
	db          *mongo.Database
	bookingColl *mongo.Collection
	redisClient *redis.Client
	seatService *SeatService
}

// NewBookingService creates a new booking service
func NewBookingService(db *mongo.Database, redisClient *redis.Client) *BookingService {
	return &BookingService{
		db:          db,
		bookingColl: db.Collection("bookings"),
		redisClient: redisClient,
	}
}

// SetSeatService sets the seat service
func (s *BookingService) SetSeatService(seatService *SeatService) {
	s.seatService = seatService
}

// CreateBooking creates a new booking
func (s *BookingService) CreateBooking(ctx context.Context, req model.BookingRequest, userID string) (*model.Booking, error) {
	// Validate request
	if req.Platform == "" {
		return nil, errors.New("platform is required")
	}

	if req.PlatformID == "" {
		return nil, errors.New("platform ID is required")
	}

	if len(req.SeatIDs) == 0 {
		return nil, errors.New("at least one seat must be selected")
	}

	// Lock seats first
	seatReq := model.SeatLockRequest{
		Platform:   req.Platform,
		PlatformID: req.PlatformID,
		SeatIDs:    req.SeatIDs,
	}

	err := s.seatService.LockSeats(ctx, seatReq, userID)
	if err != nil {
		return nil, err
	}

	// Create booking ID
	id := primitive.NewObjectID()

	// Calculate price (in a real system, fetch from the relevant platform service)
	seatPrice := 100.0 // Default price per seat
	totalPrice := seatPrice * float64(len(req.SeatIDs))

	// Create booking
	booking := &model.Booking{
		ID:          id.Hex(),
		UserID:      userID,
		ShowID:      req.PlatformID, // Using platformID as showID
		Seats:       req.SeatIDs,
		TotalPrice:  totalPrice,
		Status:      "pending",
		BookingTime: time.Now(),
		ExpiryTime:  time.Now().Add(5 * time.Minute),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	_, err = s.bookingColl.InsertOne(ctx, booking)
	if err != nil {
		// If insert fails, release the seats
		s.seatService.ReleaseSeats(ctx, seatReq, userID)
		return nil, err
	}

	return booking, nil
}

// GetBooking retrieves a booking by ID
func (s *BookingService) GetBooking(ctx context.Context, bookingID string) (*model.Booking, error) {
	// Convert string ID to ObjectID
	id, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return nil, errors.New("invalid booking ID")
	}

	// Find booking
	var booking model.Booking
	err = s.bookingColl.FindOne(ctx, bson.M{"_id": id}).Decode(&booking)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("booking not found")
		}
		return nil, err
	}

	return &booking, nil
}

// GetUserBookings retrieves all bookings for a user
func (s *BookingService) GetUserBookings(ctx context.Context, userID string, page, pageSize int) ([]model.Booking, int64, error) {
	// Set pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	// Create filter
	filter := bson.M{"user_id": userID}

	// Count total
	total, err := s.bookingColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find bookings
	cursor, err := s.bookingColl.Find(ctx, filter, options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode bookings
	var bookings []model.Booking
	if err = cursor.All(ctx, &bookings); err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

// CancelBooking cancels a booking
func (s *BookingService) CancelBooking(ctx context.Context, bookingID string) error {
	// Get the booking first
	booking, err := s.GetBooking(ctx, bookingID)
	if err != nil {
		return err
	}

	// Only pending or confirmed bookings can be cancelled
	if booking.Status != "pending" && booking.Status != "confirmed" {
		return errors.New("booking cannot be cancelled in its current state")
	}

	// Update booking status
	id, _ := primitive.ObjectIDFromHex(bookingID)
	_, err = s.bookingColl.UpdateOne(
		ctx,
		bson.M{"_id": id},
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

	// Release seats if it was a confirmed booking
	// For pending bookings, they will be automatically released when the lock expires
	if booking.Status == "confirmed" {
		// Create a seat lock request
		seatReq := model.SeatLockRequest{
			Platform:   "unknown", // We would need to store this in the booking
			PlatformID: booking.ShowID,
			SeatIDs:    booking.Seats,
		}

		// Release the seats
		s.seatService.ReleaseSeats(ctx, seatReq, booking.UserID)
	}

	return nil
}
