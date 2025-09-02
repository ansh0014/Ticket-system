package movie

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Theater represents a movie theater
type Theater struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Address   string             `json:"address" bson:"address"`
	City      string             `json:"city" bson:"city"`
	State     string             `json:"state" bson:"state"`
	Country   string             `json:"country" bson:"country"`
	ZipCode   string             `json:"zip_code" bson:"zip_code"`
	Screens   int                `json:"screens" bson:"screens"`
	Amenities []string           `json:"amenities" bson:"amenities"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// Movie represents a movie in the system
type Movie struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	Duration    int                `json:"duration" bson:"duration"` // in minutes
	ReleaseDate time.Time          `json:"release_date" bson:"release_date"`
	EndDate     time.Time          `json:"end_date" bson:"end_date"`
	Language    string             `json:"language" bson:"language"`
	Genres      []string           `json:"genres" bson:"genres"`
	Cast        []string           `json:"cast" bson:"cast"`
	Director    string             `json:"director" bson:"director"`
	Rating      string             `json:"rating" bson:"rating"` // PG, PG-13, R, etc.
	PosterImage string             `json:"poster_image" bson:"poster_image"`
	BannerImage string             `json:"banner_image" bson:"banner_image"`
	TrailerURL  string             `json:"trailer_url" bson:"trailer_url"`
	Status      string             `json:"status" bson:"status"` // upcoming, now_showing, ended
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Screen represents a screen in a theater
type Screen struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TheaterID  primitive.ObjectID `json:"theater_id" bson:"theater_id"`
	Name       string             `json:"name" bson:"name"`
	Capacity   int                `json:"capacity" bson:"capacity"`
	ScreenType string             `json:"screen_type" bson:"screen_type"` // standard, imax, 3d, etc.
	SeatLayout [][]string         `json:"seat_layout" bson:"seat_layout"` // 2D array of seat IDs
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

// Show represents a movie show
type Show struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	MovieID     primitive.ObjectID `json:"movie_id" bson:"movie_id"`
	Movie       Movie              `json:"movie" bson:"movie,omitempty"`
	TheaterID   primitive.ObjectID `json:"theater_id" bson:"theater_id"`
	Theater     Theater            `json:"theater" bson:"theater,omitempty"`
	ScreenID    primitive.ObjectID `json:"screen_id" bson:"screen_id"`
	Screen      Screen             `json:"screen" bson:"screen,omitempty"`
	StartTime   time.Time          `json:"start_time" bson:"start_time"`
	EndTime     time.Time          `json:"end_time" bson:"end_time"`
	Date        time.Time          `json:"date" bson:"date"`
	Language    string             `json:"language" bson:"language"`
	Format      string             `json:"format" bson:"format"` // 2D, 3D, IMAX, etc.
	BasePrice   float64            `json:"base_price" bson:"base_price"`
	TotalSeats  int                `json:"total_seats" bson:"total_seats"`
	AvailSeats  int                `json:"avail_seats" bson:"avail_seats"`
	BookedSeats int                `json:"booked_seats" bson:"booked_seats"`
	Status      string             `json:"status" bson:"status"` // active, cancelled, completed
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// ShowSeat represents a seat for a movie show
type ShowSeat struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ShowID      primitive.ObjectID `json:"show_id" bson:"show_id"`
	SeatNumber  string             `json:"seat_number" bson:"seat_number"`
	Row         string             `json:"row" bson:"row"`
	Column      int                `json:"column" bson:"column"`
	Category    string             `json:"category" bson:"category"` // standard, premium, recliner
	Price       float64            `json:"price" bson:"price"`
	IsAvailable bool               `json:"is_available" bson:"is_available"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// SearchMoviesRequest represents a request to search for movies
type SearchMoviesRequest struct {
	Query    string   `json:"query"`
	City     string   `json:"city"`
	Genres   []string `json:"genres"`
	Language string   `json:"language"`
	Date     string   `json:"date"`
}

// MovieResponse represents a movie with additional formatted data
type MovieResponse struct {
	Movie          *Movie `json:"movie"`
	DurationStr    string `json:"duration_str"`
	ReleaseDateStr string `json:"release_date_str"`
	GenresStr      string `json:"genres_str"`
}
