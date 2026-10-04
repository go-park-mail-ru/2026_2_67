package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"bmstuozon/backend/internal/auth"
	"bmstuozon/backend/internal/common"
	"bmstuozon/backend/internal/storage"

	"github.com/joho/godotenv"
)

type UserHandler struct {
	storage *storage.InMemoryDB
}

type ProfileResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Profile реализует /api/v1/users/{id} (!ВРЕМЕННО)
func (h *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	if idString == "" {
		http.Error(w, "ID пользователя не указан", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		http.Error(w, "Невалидный ID", http.StatusBadRequest)
		return
	}

	user, ok := h.storage.SelectUserByID(id)
	if !ok {
		http.Error(w, "Невалидный ID", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(ProfileResponse{
		user.Login, user.Email,
	})
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		return
	}
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	jwtIssuer := os.Getenv("JWT_ISSUER")

	storage := storage.NewInMemoryDB()

	authHandler := auth.NewAuthHandler(jwtSecret, jwtIssuer, storage)
	userHandler := UserHandler{
		storage,
	}

	mux := http.NewServeMux()

	// Login
	loginHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Login),
	)
	mux.Handle("/api/v1/auth/login", loginHandler)

	// Register
	registerHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Register),
	)
	mux.Handle("/api/v1/auth/register", registerHandler)

	// Refresh
	refreshHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Refresh),
	)
	mux.Handle("/api/v1/auth/refresh", refreshHandler)

	// Logout
	logoutHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Logout),
	)
	mux.Handle("/api/v1/auth/logout", logoutHandler)

	// UserProfile
	userProfileHandler := common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(userHandler.Profile),
	)
	mux.Handle("/api/v1/users/{id}", userProfileHandler)

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
