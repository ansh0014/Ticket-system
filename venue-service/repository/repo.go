package repository

import (
    "context"
    "errors"
    "time"

    "github.com/ansh0014/venue/model"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
    db        *mongo.Database
    venuesCol *mongo.Collection
    hallsCol  *mongo.Collection
    seatsCol  *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
    return &Repository{
        db:        db,
        venuesCol: db.Collection("venues"),
        hallsCol:  db.Collection("halls"),
        seatsCol:  db.Collection("seats"),
    }
}

// Venues
func (r *Repository) CreateVenue(ctx context.Context, v *model.Venue) (*model.Venue, error) {
    now := time.Now().UTC()
    v.CreatedAt = now
    v.UpdatedAt = now
    res, err := r.venuesCol.InsertOne(ctx, v)
    if err != nil {
        return nil, err
    }
    v.ID = res.InsertedID.(primitive.ObjectID)
    return v, nil
}

func (r *Repository) GetVenueByID(ctx context.Context, id primitive.ObjectID) (*model.Venue, error) {
    var v model.Venue
    if err := r.venuesCol.FindOne(ctx, bson.M{"_id": id}).Decode(&v); err != nil {
        return nil, err
    }
    return &v, nil
}

func (r *Repository) ListVenues(ctx context.Context, filter bson.M, page, pageSize int) ([]model.Venue, int64, error) {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 20
    }
    skip := int64((page - 1) * pageSize)
    limit := int64(pageSize)
    opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{"created_at", -1}})
    cur, err := r.venuesCol.Find(ctx, filter, opts)
    if err != nil {
        return nil, 0, err
    }
    defer cur.Close(ctx)
    var out []model.Venue
    if err := cur.All(ctx, &out); err != nil {
        return nil, 0, err
    }
    count, _ := r.venuesCol.CountDocuments(ctx, filter)
    return out, count, nil
}

// Halls
func (r *Repository) CreateHall(ctx context.Context, h *model.Hall) (*model.Hall, error) {
    now := time.Now().UTC()
    h.CreatedAt = now
    h.UpdatedAt = now
    res, err := r.hallsCol.InsertOne(ctx, h)
    if err != nil {
        return nil, err
    }
    h.ID = res.InsertedID.(primitive.ObjectID)
    return h, nil
}

func (r *Repository) GetHallByID(ctx context.Context, id primitive.ObjectID) (*model.Hall, error) {
    var h model.Hall
    if err := r.hallsCol.FindOne(ctx, bson.M{"_id": id}).Decode(&h); err != nil {
        return nil, err
    }
    return &h, nil
}

func (r *Repository) ListHallsByVenue(ctx context.Context, venueID primitive.ObjectID) ([]model.Hall, error) {
    cur, err := r.hallsCol.Find(ctx, bson.M{"venue_id": venueID})
    if err != nil {
        return nil, err
    }
    defer cur.Close(ctx)
    var out []model.Hall
    if err := cur.All(ctx, &out); err != nil {
        return nil, err
    }
    return out, nil
}

// Seats
func (r *Repository) AddSeat(ctx context.Context, s *model.Seat) (*model.Seat, error) {
    now := time.Now().UTC()
    s.CreatedAt = now
    s.UpdatedAt = now
    if s.IsActive == false {
        s.IsActive = true
    }
    res, err := r.seatsCol.InsertOne(ctx, s)
    if err != nil {
        return nil, err
    }
    s.ID = res.InsertedID.(primitive.ObjectID)
    return s, nil
}

func (r *Repository) ListSeatsByHall(ctx context.Context, hallID primitive.ObjectID) ([]model.Seat, error) {
    cur, err := r.seatsCol.Find(ctx, bson.M{"hall_id": hallID})
    if err != nil {
        return nil, err
    }
    defer cur.Close(ctx)
    var out []model.Seat
    if err := cur.All(ctx, &out); err != nil {
        return nil, err
    }
    return out, nil
}

func (r *Repository) UpdateSeatAvailability(ctx context.Context, seatID primitive.ObjectID, active bool) error {
    res, err := r.seatsCol.UpdateOne(ctx, bson.M{"_id": seatID}, bson.M{"$set": bson.M{"is_active": active, "updated_at": time.Now().UTC()}})
    if err != nil {
        return err
    }
    if res.MatchedCount == 0 {
        return errors.New("seat not found")
    }
    return nil
}