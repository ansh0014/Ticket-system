package event

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// Repository handles event data access
type Repository struct {
    db             *mongo.Database
    eventsColl     *mongo.Collection
    venuesColl     *mongo.Collection
    organizersColl *mongo.Collection
    categoriesColl *mongo.Collection
    seatsColl      *mongo.Collection
    ticketsColl    *mongo.Collection
}

// NewRepository creates a new event repository
func NewRepository(db *mongo.Database) *Repository {
    return &Repository{
        db:             db,
        eventsColl:     db.Collection("events"),
        venuesColl:     db.Collection("venues"),
        organizersColl: db.Collection("organizers"),
        categoriesColl: db.Collection("categories"),
        seatsColl:      db.Collection("event_seats"),
        ticketsColl:    db.Collection("ticket_types"),
    }
}

// SearchEvents searches for events based on criteria
func (r *Repository) SearchEvents(ctx context.Context, search SearchEventsRequest, page, pageSize int) ([]Event, int64, error) {
    filter := bson.M{}

    // Add query filters
    if search.Query != "" {
        filter["$or"] = []bson.M{
            {"title": bson.M{"$regex": search.Query, "$options": "i"}},
            {"description": bson.M{"$regex": search.Query, "$options": "i"}},
        }
    }

    // Filter by city
    if search.City != "" {
        // First, find venues in the city
        venueFilter := bson.M{"city": bson.M{"$regex": search.City, "$options": "i"}}
        venueCursor, err := r.venuesColl.Find(ctx, venueFilter)
        if err != nil {
            return nil, 0, err
        }

        var venues []Venue
        if err = venueCursor.All(ctx, &venues); err != nil {
            return nil, 0, err
        }

        venueIDs := make([]primitive.ObjectID, len(venues))
        for i, venue := range venues {
            venueIDs[i] = venue.ID
        }

        if len(venueIDs) > 0 {
            filter["venue_id"] = bson.M{"$in": venueIDs}
        } else {
            // No venues found in the city, return empty result
            return []Event{}, 0, nil
        }
    }

    // Filter by category
    if search.Category != "" {
        // Find category by name
        categoryFilter := bson.M{"name": bson.M{"$regex": search.Category, "$options": "i"}}
        var category Category
        err := r.categoriesColl.FindOne(ctx, categoryFilter).Decode(&category)
        if err != nil {
            if err == mongo.ErrNoDocuments {
                // No category found, return empty result
                return []Event{}, 0, nil
            }
            return nil, 0, err
        }

        filter["category_ids"] = category.ID
    }

    // Filter by date range
    if !search.StartDate.IsZero() {
        if filter["start_time"] == nil {
            filter["start_time"] = bson.M{}
        }
        filter["start_time"].(bson.M)["$gte"] = search.StartDate
    }

    if !search.EndDate.IsZero() {
        if filter["end_time"] == nil {
            filter["end_time"] = bson.M{}
        }
        filter["end_time"].(bson.M)["$lte"] = search.EndDate
    }

    // Filter by price range
    if search.PriceMin > 0 || search.PriceMax > 0 {
        // This is a bit complex as price is in ticket_types
        // We'll need to find events with ticket types in the price range
        var ticketTypeFilter bson.M
        if search.PriceMin > 0 && search.PriceMax > 0 {
            ticketTypeFilter = bson.M{"price": bson.M{"$gte": search.PriceMin, "$lte": search.PriceMax}}
        } else if search.PriceMin > 0 {
            ticketTypeFilter = bson.M{"price": bson.M{"$gte": search.PriceMin}}
        } else {
            ticketTypeFilter = bson.M{"price": bson.M{"$lte": search.PriceMax}}
        }

        // Find ticket types in the price range
        ticketCursor, err := r.ticketsColl.Find(ctx, ticketTypeFilter)
        if err != nil {
            return nil, 0, err
        }

        var ticketTypes []TicketType
        if err = ticketCursor.All(ctx, &ticketTypes); err != nil {
            return nil, 0, err
        }

        eventIDs := make([]primitive.ObjectID, len(ticketTypes))
        for i, ticketType := range ticketTypes {
            eventIDs[i] = ticketType.EventID
        }

        if len(eventIDs) > 0 {
            filter["_id"] = bson.M{"$in": eventIDs}
        } else {
            // No events found in the price range, return empty result
            return []Event{}, 0, nil
        }
    }

    // Filter by ticket availability
    if search.TicketCount > 0 {
        filter["available_seats"] = bson.M{"$gte": search.TicketCount}
    }

    // Only show upcoming events by default
    if filter["status"] == nil {
        filter["status"] = "upcoming"
    }

    // Count total matching events
    total, err := r.eventsColl.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, err
    }

    // Set up pagination
    skip := int64((page - 1) * pageSize)
    limit := int64(pageSize)

    opts := options.Find().
        SetSkip(skip).
        SetLimit(limit).
        SetSort(bson.D{{"start_time", 1}})

    // Execute the query
    cursor, err := r.eventsColl.Find(ctx, filter, opts)
    if err != nil {
        return nil, 0, err
    }
    defer cursor.Close(ctx)

    var events []Event
    if err = cursor.All(ctx, &events); err != nil {
        return nil, 0, err
    }

    // Populate related data for each event
    for i := range events {
        // Populate venue
        venue, err := r.getVenueByID(ctx, events[i].VenueID)
        if err == nil {
            events[i].Venue = *venue
        }

        // Populate organizer
        organizer, err := r.getOrganizerByID(ctx, events[i].OrganizerID)
        if err == nil {
            events[i].Organizer = *organizer
        }

        // Populate categories
        if len(events[i].CategoryIDs) > 0 {
            categories, err := r.getCategoriesByIDs(ctx, events[i].CategoryIDs)
            if err == nil {
                events[i].Categories = categories
            }
        }

        // Populate ticket types
        ticketTypes, err := r.getTicketTypesByEventID(ctx, events[i].ID)
        if err == nil {
            events[i].TicketTypes = ticketTypes
        }
    }

    return events, total, nil
}

// GetEventByID retrieves an event by ID
func (r *Repository) GetEventByID(ctx context.Context, id primitive.ObjectID) (*Event, error) {
    var event Event
    err := r.eventsColl.FindOne(ctx, bson.M{"_id": id}).Decode(&event)
    if err != nil {
        return nil, err
    }

    // Populate venue
    venue, err := r.getVenueByID(ctx, event.VenueID)
    if err == nil {
        event.Venue = *venue
    }

    // Populate organizer
    organizer, err := r.getOrganizerByID(ctx, event.OrganizerID)
    if err == nil {
        event.Organizer = *organizer
    }

    // Populate categories
    if len(event.CategoryIDs) > 0 {
        categories, err := r.getCategoriesByIDs(ctx, event.CategoryIDs)
        if err == nil {
            event.Categories = categories
        }
    }

    // Populate ticket types
    ticketTypes, err := r.getTicketTypesByEventID(ctx, event.ID)
    if err == nil {
        event.TicketTypes = ticketTypes
    }

    return &event, nil
}

// GetEventSeats retrieves seats for an event
func (r *Repository) GetEventSeats(ctx context.Context, eventID primitive.ObjectID, ticketTypeID primitive.ObjectID) ([]EventSeat, error) {
    filter := bson.M{"event_id": eventID}

    // If ticket type ID is provided, filter by it
    if !ticketTypeID.IsZero() {
        filter["ticket_type_id"] = ticketTypeID
    }

    cursor, err := r.seatsColl.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var seats []EventSeat
    if err = cursor.All(ctx, &seats); err != nil {
        return nil, err
    }

    return seats, nil
}

// GetTicketTypes retrieves ticket types for an event
func (r *Repository) GetTicketTypes(ctx context.Context, eventID primitive.ObjectID) ([]TicketType, error) {
    return r.getTicketTypesByEventID(ctx, eventID)
}

// LockEventSeats locks seats for an event
func (r *Repository) LockEventSeats(ctx context.Context, eventID primitive.ObjectID, ticketTypeID primitive.ObjectID, seatIDs []primitive.ObjectID) error {
    filter := bson.M{
        "event_id":     eventID,
        "_id":          bson.M{"$in": seatIDs},
        "is_available": true,
    }

    // If ticket type ID is provided, add it to the filter
    if !ticketTypeID.IsZero() {
        filter["ticket_type_id"] = ticketTypeID
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
            "event_id": eventID,
            "_id":      bson.M{"$in": seatIDs},
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

    // Update available seats count on the event
    _, err = r.eventsColl.UpdateOne(
        ctx,
        bson.M{"_id": eventID},
        bson.M{
            "$inc": bson.M{"available_seats": -len(seatIDs)},
            "$set": bson.M{"updated_at": time.Now()},
        },
    )

    return err
}

// UnlockEventSeats unlocks previously locked seats
func (r *Repository) UnlockEventSeats(ctx context.Context, eventID primitive.ObjectID, seatIDs []primitive.ObjectID) error {
    filter := bson.M{
        "event_id": eventID,
        "_id":      bson.M{"$in": seatIDs},
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

    // Update available seats count on the event
    _, err = r.eventsColl.UpdateOne(
        ctx,
        bson.M{"_id": eventID},
        bson.M{
            "$inc": bson.M{"available_seats": len(seatIDs)},
            "$set": bson.M{"updated_at": time.Now()},
        },
    )

    return err
}

// Helper methods to populate related data
func (r *Repository) getVenueByID(ctx context.Context, id primitive.ObjectID) (*Venue, error) {
    var venue Venue
    err := r.venuesColl.FindOne(ctx, bson.M{"_id": id}).Decode(&venue)
    if err != nil {
        return nil, err
    }
    return &venue, nil
}

func (r *Repository) getOrganizerByID(ctx context.Context, id primitive.ObjectID) (*Organizer, error) {
    var organizer Organizer
    err := r.organizersColl.FindOne(ctx, bson.M{"_id": id}).Decode(&organizer)
    if err != nil {
        return nil, err
    }
    return &organizer, nil
}

func (r *Repository) getCategoriesByIDs(ctx context.Context, ids []primitive.ObjectID) ([]Category, error) {
    filter := bson.M{"_id": bson.M{"$in": ids}}
    cursor, err := r.categoriesColl.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var categories []Category
    if err = cursor.All(ctx, &categories); err != nil {
        return nil, err
    }

    return categories, nil
}

func (r *Repository) getTicketTypesByEventID(ctx context.Context, eventID primitive.ObjectID) ([]TicketType, error) {
    filter := bson.M{"event_id": eventID}
    cursor, err := r.ticketsColl.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var ticketTypes []TicketType
    if err = cursor.All(ctx, &ticketTypes); err != nil {
        return nil, err
    }

    return ticketTypes, nil
}

// GetSeatsByIDs retrieves seats by their IDs
func (r *Repository) GetSeatsByIDs(ctx context.Context, eventID primitive.ObjectID, seatIDs []primitive.ObjectID) ([]EventSeat, error) {
    filter := bson.M{
        "event_id": eventID,
        "_id":      bson.M{"$in": seatIDs},
    }

    cursor, err := r.seatsColl.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var seats []EventSeat
    if err = cursor.All(ctx, &seats); err != nil {
        return nil, err
    }

    return seats, nil
}