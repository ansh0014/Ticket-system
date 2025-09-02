package movie

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// Repository handles movie data access
type Repository struct {
    db            *mongo.Database
    moviesColl    *mongo.Collection
    theatersColl  *mongo.Collection
    screensColl   *mongo.Collection
    showsColl     *mongo.Collection
    seatsColl     *mongo.Collection
}

// NewRepository creates a new movie repository
func NewRepository(db *mongo.Database) *Repository {
    return &Repository{
        db:            db,
        moviesColl:    db.Collection("movies"),
        theatersColl:  db.Collection("theaters"),
        screensColl:   db.Collection("screens"),
        showsColl:     db.Collection("shows"),
        seatsColl:     db.Collection("show_seats"),
    }
}

// GetMovies retrieves a list of movies
func (r *Repository) GetMovies(ctx context.Context, page, pageSize int) ([]Movie, int64, error) {
    // Calculate skip and limit
    skip := int64((page - 1) * pageSize)
    limit := int64(pageSize)

    // Set options
    findOptions := options.Find().
        SetSkip(skip).
        SetLimit(limit).
        SetSort(bson.D{{"release_date", -1}})

    // Set filter for active movies
    filter := bson.M{
        "status": bson.M{
            "$in": []string{"upcoming", "now_showing"},
        },
    }

    // Count total
    total, err := r.moviesColl.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, err
    }

    // Execute the query
    cursor, err := r.moviesColl.Find(ctx, filter, findOptions)
    if err != nil {
        return nil, 0, err
    }
    defer cursor.Close(ctx)

    var movies []Movie
    if err = cursor.All(ctx, &movies); err != nil {
        return nil, 0, err
    }

    return movies, total, nil
}

// SearchMovies searches for movies based on criteria
func (r *Repository) SearchMovies(ctx context.Context, search SearchMoviesRequest, page, pageSize int) ([]Movie, int64, error) {
    filter := bson.M{
        "status": bson.M{
            "$in": []string{"upcoming", "now_showing"},
        },
    }

    // Add query filter
    if search.Query != "" {
        filter["$or"] = []bson.M{
            {"title": bson.M{"$regex": search.Query, "$options": "i"}},
            {"description": bson.M{"$regex": search.Query, "$options": "i"}},
            {"cast": bson.M{"$regex": search.Query, "$options": "i"}},
            {"director": bson.M{"$regex": search.Query, "$options": "i"}},
        }
    }

    // Add language filter
    if search.Language != "" {
        filter["language"] = search.Language
    }

    // Add genres filter
    if len(search.Genres) > 0 {
        filter["genres"] = bson.M{"$in": search.Genres}
    }

    // Calculate skip and limit
    skip := int64((page - 1) * pageSize)
    limit := int64(pageSize)

    // Set options
    findOptions := options.Find().
        SetSkip(skip).
        SetLimit(limit).
        SetSort(bson.D{{"release_date", -1}})

    // Count total
    total, err := r.moviesColl.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, err
    }

    // Execute the query
    cursor, err := r.moviesColl.Find(ctx, filter, findOptions)
    if err != nil {
        return nil, 0, err
    }
    defer cursor.Close(ctx)

    var movies []Movie
    if err = cursor.All(ctx, &movies); err != nil {
        return nil, 0, err
    }

    return movies, total, nil
}

// GetMovieByID retrieves a movie by ID
func (r *Repository) GetMovieByID(ctx context.Context, id primitive.ObjectID) (*Movie, error) {
    var movie Movie
    err := r.moviesColl.FindOne(ctx, bson.M{"_id": id}).Decode(&movie)
    if err != nil {
        return nil, err
    }
    return &movie, nil
}

// GetMovieShows retrieves shows for a movie
func (r *Repository) GetMovieShows(ctx context.Context, movieID primitive.ObjectID, date time.Time, theaterID primitive.ObjectID) ([]Show, error) {
    // Create filter
    filter := bson.M{
        "movie_id": movieID,
        "status": "active",
    }

    // Add date filter if provided
    if !date.IsZero() {
        // Convert date to start and end of day
        startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
        endOfDay := startOfDay.Add(24 * time.Hour)
        
        filter["date"] = bson.M{
            "$gte": startOfDay,
            "$lt":  endOfDay,
        }
    }

    // Add theater filter if provided
    if theaterID != (primitive.ObjectID{}) {
        filter["theater_id"] = theaterID
    }

    // Set options for sorting
    opts := options.Find().SetSort(bson.D{{"start_time", 1}})

    // Execute the query
    cursor, err := r.showsColl.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var shows []Show
    if err = cursor.All(ctx, &shows); err != nil {
        return nil, err
    }

    // Populate related data
    for i := range shows {
        // Populate movie
        movie, err := r.GetMovieByID(ctx, shows[i].MovieID)
        if err == nil {
            shows[i].Movie = *movie
        }

        // Populate theater
        theater, err := r.getTheaterByID(ctx, shows[i].TheaterID)
        if err == nil {
            shows[i].Theater = *theater
        }

        // Populate screen
        screen, err := r.getScreenByID(ctx, shows[i].ScreenID)
        if err == nil {
            shows[i].Screen = *screen
        }
    }

    return shows, nil
}

// GetShowByID retrieves a show by ID
func (r *Repository) GetShowByID(ctx context.Context, id primitive.ObjectID) (*Show, error) {
    var show Show
    err := r.showsColl.FindOne(ctx, bson.M{"_id": id}).Decode(&show)
    if err != nil {
        return nil, err
    }

    // Populate movie
    movie, err := r.GetMovieByID(ctx, show.MovieID)
    if err == nil {
        show.Movie = *movie
    }

    // Populate theater
    theater, err := r.getTheaterByID(ctx, show.TheaterID)
    if err == nil {
        show.Theater = *theater
    }

    // Populate screen
    screen, err := r.getScreenByID(ctx, show.ScreenID)
    if err == nil {
        show.Screen = *screen
    }

    return &show, nil
}

// GetShowSeats retrieves seats for a show
func (r *Repository) GetShowSeats(ctx context.Context, showID primitive.ObjectID) ([]ShowSeat, error) {
    filter := bson.M{"show_id": showID}

    cursor, err := r.seatsColl.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var seats []ShowSeat
    if err = cursor.All(ctx, &seats); err != nil {
        return nil, err
    }

    return seats, nil
}

// LockShowSeats locks seats for a booking
func (r *Repository) LockShowSeats(ctx context.Context, showID primitive.ObjectID, seatIDs []primitive.ObjectID) error {
    filter := bson.M{
        "show_id":      showID,
        "_id":          bson.M{"$in": seatIDs},
        "is_available": true,
    }

    update := bson.M{
        "$set": bson.M{
            "is_available": false,
            "updated_at":   time.Now(),
        },
    }

    result, err := r.seatsColl.UpdateMany(ctx, filter, update)
    if err != nil {
        return err
    }

    // Check if all seats were updated
    if result.ModifiedCount != int64(len(seatIDs)) {
        // Revert the changes
        revertFilter := bson.M{
            "show_id": showID,
            "_id":     bson.M{"$in": seatIDs},
        }

        revertUpdate := bson.M{
            "$set": bson.M{
                "is_available": true,
                "updated_at":   time.Now(),
            },
        }

        _, _ = r.seatsColl.UpdateMany(ctx, revertFilter, revertUpdate)

        return errors.New("some seats are no longer available")
    }

    // Update available seats count on the show
    _, err = r.showsColl.UpdateOne(
        ctx,
        bson.M{"_id": showID},
        bson.M{
            "$inc": bson.M{
                "avail_seats":  -len(seatIDs),
                "booked_seats": len(seatIDs),
            },
            "$set": bson.M{"updated_at": time.Now()},
        },
    )

    return err
}

// UnlockShowSeats unlocks previously locked seats
func (r *Repository) UnlockShowSeats(ctx context.Context, showID primitive.ObjectID, seatIDs []primitive.ObjectID) error {
    filter := bson.M{
        "show_id": showID,
        "_id":     bson.M{"$in": seatIDs},
    }

    update := bson.M{
        "$set": bson.M{
            "is_available": true,
            "updated_at":   time.Now(),
        },
    }

    _, err := r.seatsColl.UpdateMany(ctx, filter, update)
    if err != nil {
        return err
    }

    // Update available seats count on the show
    _, err = r.showsColl.UpdateOne(
        ctx,
        bson.M{"_id": showID},
        bson.M{
            "$inc": bson.M{
                "avail_seats":  len(seatIDs),
                "booked_seats": -len(seatIDs),
            },
            "$set": bson.M{"updated_at": time.Now()},
        },
    )

    return err
}

// Helper methods
func (r *Repository) getTheaterByID(ctx context.Context, id primitive.ObjectID) (*Theater, error) {
    var theater Theater
    err := r.theatersColl.FindOne(ctx, bson.M{"_id": id}).Decode(&theater)
    if err != nil {
        return nil, err
    }
    return &theater, nil
}

func (r *Repository) getScreenByID(ctx context.Context, id primitive.ObjectID) (*Screen, error) {
    var screen Screen
    err := r.screensColl.FindOne(ctx, bson.M{"_id": id}).Decode(&screen)
    if err != nil {
        return nil, err
    }
    return &screen, nil
}

// GetTheaters retrieves theaters with optional city filter
func (r *Repository) GetTheaters(ctx context.Context, city string, page, pageSize int) ([]Theater, int64, error) {
    filter := bson.M{}
    if city != "" {
        filter["city"] = bson.M{"$regex": city, "$options": "i"}
    }

    // Calculate skip and limit
    skip := int64((page - 1) * pageSize)
    limit := int64(pageSize)

    // Set options
    findOptions := options.Find().
        SetSkip(skip).
        SetLimit(limit).
        SetSort(bson.D{{"name", 1}})

    // Count total
    total, err := r.theatersColl.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, err
    }

    // Execute the query
    cursor, err := r.theatersColl.Find(ctx, filter, findOptions)
    if err != nil {
        return nil, 0, err
    }
    defer cursor.Close(ctx)

    var theaters []Theater
    if err = cursor.All(ctx, &theaters); err != nil {
        return nil, 0, err
    }

    return theaters, total, nil
}

// GetTheaterByID retrieves a theater by ID
func (r *Repository) GetTheaterByID(ctx context.Context, id primitive.ObjectID) (*Theater, error) {
    var theater Theater
    err := r.theatersColl.FindOne(ctx, bson.M{"_id": id}).Decode(&theater)
    if err != nil {
        return nil, err
    }
    return &theater, nil
}

// GetTheaterShows retrieves shows for a specific theater
func (r *Repository) GetTheaterShows(ctx context.Context, theaterID primitive.ObjectID, date time.Time) ([]Show, error) {
    // Create filter
    filter := bson.M{
        "theater_id": theaterID,
        "status": "active",
    }

    // Add date filter if provided
    if !date.IsZero() {
        // Convert date to start and end of day
        startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
        endOfDay := startOfDay.Add(24 * time.Hour)
        
        filter["date"] = bson.M{
            "$gte": startOfDay,
            "$lt":  endOfDay,
        }
    }

    // Set options for sorting
    opts := options.Find().SetSort(bson.D{{"start_time", 1}})

    // Execute the query
    cursor, err := r.showsColl.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var shows []Show
    if err = cursor.All(ctx, &shows); err != nil {
        return nil, err
    }

    // Populate related data
    for i := range shows {
        // Populate movie
        movie, err := r.GetMovieByID(ctx, shows[i].MovieID)
        if err == nil {
            shows[i].Movie = *movie
        }

        // Populate theater
        theater, err := r.getTheaterByID(ctx, shows[i].TheaterID)
        if err == nil {
            shows[i].Theater = *theater
        }

        // Populate screen
        screen, err := r.getScreenByID(ctx, shows[i].ScreenID)
        if err == nil {
            shows[i].Screen = *screen
        }
    }

    return shows, nil
}