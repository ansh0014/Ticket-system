package service

import (
    "context"

    "github.com/ansh0014/venue/model"
	"github.com/ansh0014/venue/repository"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
    repo *repository.Repository
}

func NewService(r *repository.Repository) *Service {
    return &Service{repo: r}
}

// Venue methods
func (s *Service) CreateVenue(ctx context.Context, v *model.Venue) (*model.Venue, error) {
    return s.repo.CreateVenue(ctx, v)
}

func (s *Service) GetVenue(ctx context.Context, id primitive.ObjectID) (*model.Venue, error) {
    return s.repo.GetVenueByID(ctx, id)
}

func (s *Service) ListVenues(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]model.Venue, int64, error) {
    bsonFilter := map[string]interface{}{}
    for k, v := range filter {
        bsonFilter[k] = v
    }
    return s.repo.ListVenues(ctx, bsonFilter, page, pageSize)
}

// Hall methods
func (s *Service) CreateHall(ctx context.Context, h *model.Hall) (*model.Hall, error) {
    return s.repo.CreateHall(ctx, h)
}

func (s *Service) GetHall(ctx context.Context, id primitive.ObjectID) (*model.Hall, error) {
    return s.repo.GetHallByID(ctx, id)
}

func (s *Service) ListHalls(ctx context.Context, venueID primitive.ObjectID) ([]model.Hall, error) {
    return s.repo.ListHallsByVenue(ctx, venueID)
}

// Seat methods
func (s *Service) AddSeat(ctx context.Context, seat *model.Seat) (*model.Seat, error) {
    return s.repo.AddSeat(ctx, seat)
}

func (s *Service) ListSeats(ctx context.Context, hallID primitive.ObjectID) ([]model.Seat, error) {
    return s.repo.ListSeatsByHall(ctx, hallID)
}

func (s *Service) SetSeatActive(ctx context.Context, seatID primitive.ObjectID, active bool) error {
    return s.repo.UpdateSeatAvailability(ctx, seatID, active)
}