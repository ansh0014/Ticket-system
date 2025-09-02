package movie

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles movie business logic
type Service struct {
	repo         *Repository
	redisClient  *redis.Client
	seatLockTime time.Duration
}

// NewService creates a new movie service
func NewService(repo *Repository, redisClient *redis.Client) *Service {
	return &Service{
		repo:         repo,
		redisClient:  redisClient,
		seatLockTime: 5 * time.Minute, // Default 5 minutes lock time
	}
}

// GetMovies retrieves a list of movies
func (s *Service) GetMovies(ctx context.Context, page, pageSize int) ([]MovieResponse, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Get movies from repository
	movies, total, err := s.repo.GetMovies(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// Format response
	var response []MovieResponse
	for _, movie := range movies {
		movieResp := formatMovieResponse(&movie)
		response = append(response, movieResp)
	}

	return response, total, nil
}

// SearchMovies searches for movies based on criteria
func (s *Service) SearchMovies(ctx context.Context, req SearchMoviesRequest, page, pageSize int) ([]MovieResponse, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Get movies from repository
	movies, total, err := s.repo.SearchMovies(ctx, req, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// Format response
	var response []MovieResponse
	for _, movie := range movies {
		movieResp := formatMovieResponse(&movie)
		response = append(response, movieResp)
	}

	return response, total, nil
}

// GetMovieByID gets detailed information about a movie
func (s *Service) GetMovieByID(ctx context.Context, movieID string) (*MovieResponse, error) {
	id, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		return nil, errors.New("invalid movie ID")
	}

	movie, err := s.repo.GetMovieByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Format the movie data
	response := formatMovieResponse(movie)
	return &response, nil
}

// GetMovieShows retrieves shows for a movie
func (s *Service) GetMovieShows(ctx context.Context, movieID string, date time.Time, theaterID string) ([]Show, error) {
	id, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		return nil, errors.New("invalid movie ID")
	}

	// Parse theater ID if provided
	var theaterObjID primitive.ObjectID
	if theaterID != "" {
		theaterObjID, err = primitive.ObjectIDFromHex(theaterID)
		if err != nil {
			return nil, errors.New("invalid theater ID")
		}
	}

	return s.repo.GetMovieShows(ctx, id, date, theaterObjID)
}

// GetShowByID retrieves details of a specific show
func (s *Service) GetShowByID(ctx context.Context, showID string) (*Show, error) {
	id, err := primitive.ObjectIDFromHex(showID)
	if err != nil {
		return nil, errors.New("invalid show ID")
	}

	return s.repo.GetShowByID(ctx, id)
}

// GetShowSeats retrieves seats for a show
func (s *Service) GetShowSeats(ctx context.Context, showID string) ([]ShowSeat, error) {
	id, err := primitive.ObjectIDFromHex(showID)
	if err != nil {
		return nil, errors.New("invalid show ID")
	}

	return s.repo.GetShowSeats(ctx, id)
}

// LockShowSeats temporarily locks seats for a booking
func (s *Service) LockShowSeats(ctx context.Context, showID string, seatIDs []string, userID string) error {
	if showID == "" {
		return errors.New("show ID is required")
	}

	if len(seatIDs) == 0 {
		return errors.New("at least one seat must be selected")
	}

	if userID == "" {
		return errors.New("user ID is required")
	}

	// Convert string IDs to ObjectIDs
	_, err := primitive.ObjectIDFromHex(showID)
	if err != nil {
		return errors.New("invalid show ID")
	}

	seatObjIDs := make([]primitive.ObjectID, 0, len(seatIDs))
	for _, id := range seatIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return errors.New("invalid seat ID: " + id)
		}
		seatObjIDs = append(seatObjIDs, objID)
	}

	// Check if seats are already locked in Redis
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("show_seat_lock:%s:%s", showID, seatID)

		// Check if the lock exists
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val != "" && val != userID {
			// Seat is locked by someone else
			return errors.New("one or more selected seats are no longer available")
		}
	}

	// Set Redis locks
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("show_seat_lock:%s:%s", showID, seatID)

		// Set the lock with expiration
		err := s.redisClient.Set(ctx, lockKey, userID, s.seatLockTime).Err()
		if err != nil {
			// If any lock fails, clean up and return error
			s.unlockSeats(ctx, showID, seatIDs, userID)
			return errors.New("failed to lock seats: " + err.Error())
		}
	}

	return nil
}

// unlockSeats helper to remove Redis locks
func (s *Service) unlockSeats(ctx context.Context, showID string, seatIDs []string, userID string) {
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("show_seat_lock:%s:%s", showID, seatID)

		// Only remove if the lock belongs to this user
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val == userID {
			s.redisClient.Del(ctx, lockKey)
		}
	}
}

// ConfirmSeats permanently reserves seats after payment
func (s *Service) ConfirmSeats(ctx context.Context, showID string, seatIDs []string) error {
	// Convert string IDs to ObjectIDs
	showObjID, err := primitive.ObjectIDFromHex(showID)
	if err != nil {
		return errors.New("invalid show ID")
	}

	seatObjIDs := make([]primitive.ObjectID, 0, len(seatIDs))
	for _, id := range seatIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return errors.New("invalid seat ID: " + id)
		}
		seatObjIDs = append(seatObjIDs, objID)
	}

	// Remove Redis locks
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("show_seat_lock:%s:%s", showID, seatID)
		s.redisClient.Del(ctx, lockKey)
	}

	// Update the database
	return s.repo.LockShowSeats(ctx, showObjID, seatObjIDs)
}

// ReleaseSeats releases previously locked seats
func (s *Service) ReleaseSeats(ctx context.Context, showID string, seatIDs []string, userID string) error {
	// Convert string IDs to ObjectIDs
	showObjID, err := primitive.ObjectIDFromHex(showID)
	if err != nil {
		return errors.New("invalid show ID")
	}

	seatObjIDs := make([]primitive.ObjectID, 0, len(seatIDs))
	for _, id := range seatIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return errors.New("invalid seat ID: " + id)
		}
		seatObjIDs = append(seatObjIDs, objID)
	}

	// Remove Redis locks
	for _, seatID := range seatIDs {
		lockKey := fmt.Sprintf("show_seat_lock:%s:%s", showID, seatID)

		// Only remove if the lock belongs to this user
		val, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == nil && val == userID {
			s.redisClient.Del(ctx, lockKey)
		}
	}

	// Update the database if these were permanent locks
	return s.repo.UnlockShowSeats(ctx, showObjID, seatObjIDs)
}

// GetTheaters retrieves theaters with optional city filter
func (s *Service) GetTheaters(ctx context.Context, city string, page, pageSize int) ([]Theater, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	return s.repo.GetTheaters(ctx, city, page, pageSize)
}

// GetTheaterByID retrieves a theater by ID
func (s *Service) GetTheaterByID(ctx context.Context, theaterID string) (*Theater, error) {
	id, err := primitive.ObjectIDFromHex(theaterID)
	if err != nil {
		return nil, errors.New("invalid theater ID")
	}

	return s.repo.GetTheaterByID(ctx, id)
}

// GetTheaterShows retrieves shows for a specific theater
func (s *Service) GetTheaterShows(ctx context.Context, theaterID string, date time.Time) ([]Show, error) {
	id, err := primitive.ObjectIDFromHex(theaterID)
	if err != nil {
		return nil, errors.New("invalid theater ID")
	}

	return s.repo.GetTheaterShows(ctx, id, date)
}

// Helper functions
func formatMovieResponse(movie *Movie) MovieResponse {
	// Format duration
	hours := movie.Duration / 60
	minutes := movie.Duration % 60
	durationStr := fmt.Sprintf("%dh %dm", hours, minutes)

	// Format release date
	releaseDateStr := movie.ReleaseDate.Format("Jan 2, 2006")

	// Format genres
	genresStr := strings.Join(movie.Genres, ", ")

	return MovieResponse{
		Movie:          movie,
		DurationStr:    durationStr,
		ReleaseDateStr: releaseDateStr,
		GenresStr:      genresStr,
	}
}
