package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ansh0014/auth/model"
	"github.com/ansh0014/auth/service"
	"github.com/ansh0014/auth/utils"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	// Check if user exists
	if user, _ := service.FindUserByEmail(req.Email); user != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	// Create user (inactive)
	if err := service.CreateUser(req.Email, string(hash)); err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}
	// Generate and send OTP
	otp, err := utils.GenerateOTP(4)
	if err != nil {
		http.Error(w, "Failed to generate OTP", http.StatusInternalServerError)
		return
	}
	if err := utils.StoreOTP(req.Email, otp, 90*time.Second); err != nil {
		http.Error(w, "Failed to store OTP", http.StatusInternalServerError)
		return
	}
	service.SendOTP(req.Email, otp)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OTP sent to email"))
}

func VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	otp, err := utils.GetOTP(req.Email)
	if err != nil {
		http.Error(w, "OTP expired or not found", http.StatusUnauthorized)
		return
	}
	if otp != req.OTP {
		http.Error(w, "Invalid OTP", http.StatusUnauthorized)
		return
	}
	utils.DeleteOTP(req.Email)
	service.ActivateUser(req.Email)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OTP verified! You can now login."))
}
