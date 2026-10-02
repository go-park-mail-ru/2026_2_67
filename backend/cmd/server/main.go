package main

import (
	"fmt"
	"net/http"

	"myapp/backend/internal/auth"
	"myapp/backend/internal/common"
)

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.UserClaimsKey).(*auth.CustomClaims)

	if !ok {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	fmt.Println("Has Claims:", ok)
	fmt.Printf("Claims: %+v\n", claims)

	w.Write([]byte("Check console"))
}

func main() {
	secret := []byte("OKAK")

	// 2. Оборачиваем обработчик в middleware
	jwtCheck := common.RequireHTTPMethod(http.MethodGet)(auth.JWTAccessTokenMiddleware(secret)(http.HandlerFunc(EchoHandler)))

	http.Handle("/", jwtCheck)

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", jwtCheck); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
