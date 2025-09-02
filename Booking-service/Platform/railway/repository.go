package railway

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository handles railway data access
type Repository struct {
	db           *mongo.Database
	trainsColl   *mongo.Collection
	stationsColl *mongo.Collection
	stopsColl    *mongo.Collection
	seatsColl    *mongo.Collection
}

// NewRepository creates a new railway repository
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		db:           db,
		trainsColl:   db.Collection("trains"),
		stationsColl: db.Collection("stations"),
		stopsColl:    db.Collection("train_stops"),
		seatsColl:    db.Collection("train_seats"),
	}
}

// GetStations retrieves a list of stations with optional city filter
func (r *Repository) GetStations(ctx context.Context, city string, page, pageSize int) ([]Station, int64, error) {
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
	total, err := r.stationsColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Execute the query
	cursor, err := r.stationsColl.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var stations []Station
	if err = cursor.All(ctx, &stations); err != nil {
		return nil, 0, err
	}

	return stations, total, nil
}

// GetStationByID retrieves a station by ID
func (r *Repository) GetStationByID(ctx context.Context, id primitive.ObjectID) (*Station, error) {
	var station Station
	err := r.stationsColl.FindOne(ctx, bson.M{"_id": id}).Decode(&station)
	if err != nil {
		return nil, err
	}
	return &station, nil
}

// GetStationByCode retrieves a station by its code
func (r *Repository) GetStationByCode(ctx context.Context, code string) (*Station, error) {
	var station Station
	err := r.stationsColl.FindOne(ctx, bson.M{"code": code}).Decode(&station)
	if err != nil {
		return nil, err
	}
	return &station, nil
}

// SearchTrains searches for trains based on criteria
func (r *Repository) SearchTrains(ctx context.Context, search SearchTrainsRequest) ([]Train, error) {
	// Get origin and destination stations first
	originStation, err := r.GetStationByCode(ctx, search.Origin)
	if err != nil {
		return nil, errors.New("origin station not found")
	}

	destStation, err := r.GetStationByCode(ctx, search.Destination)
	if err != nil {
		return nil, errors.New("destination station not found")
	}

	// Get the day of the week for the search date
	dayOfWeek := search.Date.Weekday().String()

	// Create a filter for trains from origin to destination on the given day
	filter := bson.M{
		"origin_id":       originStation.ID,
		"destination_id":  destStation.ID,
		"frequency":       dayOfWeek,
		"available_seats": bson.M{"$gte": search.Passengers},
		"status":          "scheduled",
	}

	// Add class filter if specified
	if search.Class != "" {
		filter["classes"] = search.Class
	}

	// Set options for sorting
	opts := options.Find().SetSort(bson.D{{"departure_time", 1}})

	// Execute the query
	cursor, err := r.trainsColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var trains []Train
	if err = cursor.All(ctx, &trains); err != nil {
		return nil, err
	}

	// Populate origin and destination for each train
	for i := range trains {
		// Populate origin
		origin, err := r.GetStationByID(ctx, trains[i].OriginID)
		if err == nil {
			trains[i].Origin = *origin
		}

		// Populate destination
		destination, err := r.GetStationByID(ctx, trains[i].DestinationID)
		if err == nil {
			trains[i].Destination = *destination
		}
	}

	return trains, nil
}

// GetTrainByID retrieves a train by ID
func (r *Repository) GetTrainByID(ctx context.Context, id primitive.ObjectID) (*Train, error) {
	var train Train
	err := r.trainsColl.FindOne(ctx, bson.M{"_id": id}).Decode(&train)
	if err != nil {
		return nil, err
	}

	// Populate origin
	origin, err := r.GetStationByID(ctx, train.OriginID)
	if err == nil {
		train.Origin = *origin
	}

	// Populate destination
	destination, err := r.GetStationByID(ctx, train.DestinationID)
	if err == nil {
		train.Destination = *destination
	}

	return &train, nil
}

// GetTrainSeats retrieves seats for a train
func (r *Repository) GetTrainSeats(ctx context.Context, trainID primitive.ObjectID, class string) ([]TrainSeat, error) {
	filter := bson.M{"train_id": trainID}

	// Add class filter if specified
	if class != "" {
		filter["class"] = class
	}

	cursor, err := r.seatsColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var seats []TrainSeat
	if err = cursor.All(ctx, &seats); err != nil {
		return nil, err
	}

	return seats, nil
}

// GetTrainStops retrieves stops for a train
func (r *Repository) GetTrainStops(ctx context.Context, trainID primitive.ObjectID) ([]TrainStop, error) {
	filter := bson.M{"train_id": trainID}

	// Set options for sorting by stop number
	opts := options.Find().SetSort(bson.D{{"stop_number", 1}})

	cursor, err := r.stopsColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stops []TrainStop
	if err = cursor.All(ctx, &stops); err != nil {
		return nil, err
	}

	// Populate station for each stop
	for i := range stops {
		station, err := r.GetStationByID(ctx, stops[i].StationID)
		if err == nil {
			stops[i].Station = *station
		}
	}

	return stops, nil
}

// LockTrainSeats locks seats for a booking
func (r *Repository) LockTrainSeats(ctx context.Context, trainID primitive.ObjectID, seatIDs []primitive.ObjectID) error {
	filter := bson.M{
		"train_id":     trainID,
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
			"train_id": trainID,
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

	// Update available seats count on the train
	_, err = r.trainsColl.UpdateOne(
		ctx,
		bson.M{"_id": trainID},
		bson.M{
			"$inc": bson.M{"available_seats": -len(seatIDs)},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)

	return err
}

// UnlockTrainSeats unlocks previously locked seats
func (r *Repository) UnlockTrainSeats(ctx context.Context, trainID primitive.ObjectID, seatIDs []primitive.ObjectID) error {
	filter := bson.M{
		"train_id": trainID,
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

	// Update available seats count on the train
	_, err = r.trainsColl.UpdateOne(
		ctx,
		bson.M{"_id": trainID},
		bson.M{
			"$inc": bson.M{"available_seats": len(seatIDs)},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)

	return err
}

// SearchTrainsByStations searches for trains that pass through both origin and destination stations
func (r *Repository) SearchTrainsByStations(ctx context.Context, originCode, destinationCode string, date time.Time) ([]Train, error) {
	// Get origin and destination stations
	originStation, err := r.GetStationByCode(ctx, originCode)
	if err != nil {
		return nil, errors.New("origin station not found")
	}

	destStation, err := r.GetStationByCode(ctx, destinationCode)
	if err != nil {
		return nil, errors.New("destination station not found")
	}

	// Get day of week
	dayOfWeek := date.Weekday().String()

	// First find all trains that have both stations as stops
	// Find train IDs where the origin station is a stop
	originPipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", originStation.ID}}}},
		bson.D{{"$group", bson.D{{"_id", "$train_id"}}}},
	}

	originCursor, err := r.stopsColl.Aggregate(ctx, originPipeline)
	if err != nil {
		return nil, err
	}
	defer originCursor.Close(ctx)

	var originResults []bson.M
	if err = originCursor.All(ctx, &originResults); err != nil {
		return nil, err
	}

	// Find train IDs where the destination station is a stop
	destPipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{{"station_id", destStation.ID}}}},
		bson.D{{"$group", bson.D{{"_id", "$train_id"}}}},
	}

	destCursor, err := r.stopsColl.Aggregate(ctx, destPipeline)
	if err != nil {
		return nil, err
	}
	defer destCursor.Close(ctx)

	var destResults []bson.M
	if err = destCursor.All(ctx, &destResults); err != nil {
		return nil, err
	}

	// Find intersection of train IDs
	trainIDs := make([]primitive.ObjectID, 0)
	originTrainIDMap := make(map[string]bool)

	for _, result := range originResults {
		trainID := result["_id"].(primitive.ObjectID)
		originTrainIDMap[trainID.Hex()] = true
	}

	for _, result := range destResults {
		trainID := result["_id"].(primitive.ObjectID)
		if originTrainIDMap[trainID.Hex()] {
			trainIDs = append(trainIDs, trainID)
		}
	}

	if len(trainIDs) == 0 {
		return []Train{}, nil
	}

	// Now get the full train details and ensure the origin stop comes before destination stop
	var validTrains []Train

	for _, trainID := range trainIDs {
		// Get all stops for this train
		stops, err := r.GetTrainStops(ctx, trainID)
		if err != nil {
			continue
		}

		// Find origin and destination stop numbers
		var originStopNum, destStopNum int
		originFound, destFound := false, false

		for _, stop := range stops {
			if stop.StationID == originStation.ID {
				originStopNum = stop.StopNumber
				originFound = true
			}
			if stop.StationID == destStation.ID {
				destStopNum = stop.StopNumber
				destFound = true
			}
		}

		// Only include trains where origin comes before destination
		if originFound && destFound && originStopNum < destStopNum {
			// Get the full train details
			train, err := r.GetTrainByID(ctx, trainID)
			if err != nil {
				continue
			}

			// Check if train operates on this day
			operates := false
			for _, day := range train.Frequency {
				if day == dayOfWeek {
					operates = true
					break
				}
			}

			if operates && train.Status == "scheduled" && train.AvailableSeats > 0 {
				validTrains = append(validTrains, *train)
			}
		}
	}

	return validTrains, nil
}
