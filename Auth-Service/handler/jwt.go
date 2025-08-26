package handler

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/ansh0014/auth/model"
	"github.com/ansh0014/auth/service"
	"github.com/ansh0014/auth/utils"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	user, err := service.FindUserByEmail(req.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	if !user.IsActive {
		http.Error(w, "User not activated", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}
	accessToken, _ := utils.GenerateJWT(user.ID)
	refreshToken, _ := utils.GenerateRefreshToken(user.ID)
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
