// Package auth для аутентификации/авторизации пользователей
package auth

import (
	"encoding/json"
	"net/http"
)

/*
* AccessToken:
* {
* ...header
* user_id: int
* email: string
* role: string
* }
 */

type Storage interface {
	CreateUser(u User) error
}

type AuthHandler struct {
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {

}

type LoginRequest struct {
	LoginOrEmail string `json:"login_or_email"`
	Password     string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	encoder := json.NewEncoder(w)
	claims, ok := r.Context().Value(UserClaimsKey).(*CustomClaims)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// пользователь залогинин
	if ok && claims != nil {
		w.Header().Set("Content-Type", "application/json")
		err := encoder.Encode(LoginResponse{
			AccessToken:  "ACCESS_TOKEN",  // тот же access, что и пришел
			RefreshToken: "REFRESH_TOKEN", // тот же refresh, что в базе данных
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	body := &LoginRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	/*
	* поиск пользователя по login_or_email
	* Если пользователь найден:
	* Создаём access_token и refresh_token (создаём запись в бд)
	* Если пользователь не найден:
	* Возвращаем 400
	 */
}
