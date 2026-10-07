package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"vibe_market/backend/internal/storage"
)

const loginURL string = "/api/v1/auth/login"

func usersTearUp(db *storage.InMemoryDB) {
	db.InsertUser("lg1", "email1@test.ru", "pw1")
	db.InsertUser("lg2", "email2@test.ru", "pw2")
	db.InsertUser("lg3", "email3@test.ru", "pw3")

	db.InsertUser("its_me", "its_me@yandex.ru", "mega_password")
}

func TestLoginPositive(t *testing.T) {
	db := storage.NewInMemoryDB()
	h := NewAuthHandler([]byte("SECRET"), "TEST", db)
	usersTearUp(db)

	testcases := []struct {
		name         string
		loginOrEmail string
		password     string
	}{
		{
			"Логин в точности присутствует в бд", "lg1", "pw1",
		},
		{
			"Логин написан в верхнем регистре", "LG1", "pw1",
		},
		{
			"Логин написан в смешанном регистре", "ITs_mE", "mega_password",
		},
		{
			"Логин написан с пробелами", "    \tlg2  ", "pw2",
		},
		{
			"Email в точности присутствует в бд", "email3@test.ru", "pw3",
		},
		{
			"Смешанный email", "     its_me@yandex.ru", "mega_password",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(h.Login))

	defer ts.Close()

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			rBody := map[string]string{
				"loginOrEmail": tt.loginOrEmail,
				"password":     tt.password,
			}

			jsonBytes, err := json.Marshal(rBody)
			if err != nil {
				t.Fatalf("ошибка сериализации JSON: %v", err)
			}

			res, err := http.Post(
				ts.URL+loginURL,
				"application/json",
				bytes.NewBuffer(jsonBytes),
			)
			if err != nil {
				t.Fatalf("ошибка выполнения запроса: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Fatalf("ожидался статус 200, получен %d", res.StatusCode)
			}

			if res.Header.Get("Set-Cookie") == "" {
				t.Fatalf("не установлен cookie для refreshToken")
			}

			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("ошибка чтения тела ответа: %v", err)
			}

			var resBody loginResponse
			if err := json.Unmarshal(bodyBytes, &resBody); err != nil {
				t.Fatalf("ошибка парсинга JSON-ответа: %v", err)
			}

			if resBody.AccessToken == "" {
				t.Error("accessToken отсутствует в теле ответа")
			}
		})
	}
}
