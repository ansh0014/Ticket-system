package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ansh0014/booking/Platform/movie"
	"github.com/ansh0014/booking/utils"
	"github.com/gorilla/mux"
)

// GetMoviesHandler retrieves a list of movies
func GetMoviesHandler(w http.ResponseWriter, r *http.Request) {
	movieService := r.Context().Value("movieService").(*movie.Service)

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	movies, total, err := movieService.GetMovies(r.Context(), page, pageSize)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get movies: "+err.Error())
		return
	}

	pagination := map[string]interface{}{
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"data":       movies,
		"pagination": pagination,
	})
}

// SearchMoviesHandler handles searching for movies
func SearchMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var req movie.SearchMoviesRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get movie service from context
	movieService := r.Context().Value("movieService").(*movie.Service)

	// Handle pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	movies, total, err := movieService.SearchMovies(r.Context(), req, page, pageSize)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to search movies: "+err.Error())
		return
	}

	pagination := map[string]interface{}{
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"data":       movies,
		"pagination": pagination,
	})
}

// GetMovieDetailsHandler retrieves details for a specific movie
func GetMovieDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	movieID := vars["id"]

	movieService := r.Context().Value("movieService").(*movie.Service)

	movieDetails, err := movieService.GetMovieByID(r.Context(), movieID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Movie not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    movieDetails,
	})
}

// GetMovieShowsHandler retrieves shows for a specific movie
func GetMovieShowsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	movieID := vars["id"]

	// Parse date filter
	dateStr := r.URL.Query().Get("date")
	var date time.Time
	var err error

	if dateStr != "" {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
			return
		}
	} else {
		date = time.Now()
	}

	// Get theater filter
	theaterID := r.URL.Query().Get("theater_id")

	movieService := r.Context().Value("movieService").(*movie.Service)

	shows, err := movieService.GetMovieShows(r.Context(), movieID, date, theaterID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get movie shows: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"shows":      shows,
			"count":      len(shows),
			"movie_id":   movieID,
			"date":       date.Format("2006-01-02"),
			"theater_id": theaterID,
		},
	})
}

// GetMovieSeatsHandler retrieves available seats for a movie show
func GetMovieSeatsHandler(w http.ResponseWriter, r *http.Request) {
	// Movie ID isn't actually used here since we need show ID
	showID := r.URL.Query().Get("show_id")
	if showID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "show_id query parameter is required")
		return
	}

	movieService := r.Context().Value("movieService").(*movie.Service)

	seats, err := movieService.GetShowSeats(r.Context(), showID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get show seats: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"seats":   seats,
			"count":   len(seats),
			"show_id": showID,
		},
	})
}

// LockMovieSeatsHandler temporarily reserves seats for a movie show
func LockMovieSeatsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ShowID  string   `json:"show_id"`
		SeatIDs []string `json:"seat_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get the user ID from the authenticated request
	userID := r.Context().Value("userID").(string)

	movieService := r.Context().Value("movieService").(*movie.Service)

	err := movieService.LockShowSeats(r.Context(), req.ShowID, req.SeatIDs, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to lock seats: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Seats locked successfully",
		"data": map[string]interface{}{
			"show_id":  req.ShowID,
			"seat_ids": req.SeatIDs,
		},
	})
}
