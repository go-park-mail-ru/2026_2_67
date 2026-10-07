package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"vibe_market/backend/internal/auth"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/storage"

	"vibe_market/backend/internal/products"

	"github.com/joho/godotenv"
)

type usersHandler struct {
	storage *storage.InMemoryDB
}

type profileResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Profile реализует /api/v1/users/{id} (!ВРЕМЕННО)
func (h *usersHandler) Profile(w http.ResponseWriter, r *http.Request) {
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
	err = encoder.Encode(profileResponse{
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
	userHandler := usersHandler{
		storage,
	}

	mux := http.NewServeMux()

	// POST /api/v1/auth/login
	loginHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Login),
	)
	mux.Handle("/api/v1/auth/login", loginHandler)

	// POST /api/v1/auth/register
	registerHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Register),
	)
	mux.Handle("/api/v1/auth/register", registerHandler)

	// POST /api/v1/auth/refresh
	refreshHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Refresh),
	)
	mux.Handle("/api/v1/auth/refresh", refreshHandler)

	// POST /api/v1/auth/logout
	logoutHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Logout),
	)
	mux.Handle("/api/v1/auth/logout", logoutHandler)

	// GET /api/v1/users/{}
	userProfileHandler := common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(userHandler.Profile),
	)
	mux.Handle("/api/v1/users/{id}", userProfileHandler)

	// 7. Инициализируем хранилище и обработчик товаров
	productsStorage := products.NewInMemoryDB()
	productsHandler := products.MakeProductsHandler(productsStorage)

	// 8. Регистрируем ProductsHandler для /api/v1/products
	productsChain := common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(productsHandler.GetProductsHandler),
	)
	mux.Handle("/api/v1/products", productsChain)

	// 9. Оборачиваем роутер в CORS middleware
	handler := common.CORSMiddleware(mux)

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
