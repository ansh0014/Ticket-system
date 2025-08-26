package router

import (
	"net/http"

	"github.com/ansh0014/auth/handler"
)

func SetupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", handler.RegisterHandler)
	mux.HandleFunc("/auth/verify-otp", handler.VerifyOTPHandler)
	mux.HandleFunc("/auth/login", handler.LoginHandler)           // If you have JWT login
	
	return mux
}
