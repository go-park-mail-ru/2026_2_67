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
		name string
		loginRequest
	}{
		{
			"Логин в точности присутствует в бд", loginRequest{new("lg1"), new("pw1")},
		},
		{
			"Логин написан в верхнем регистре", loginRequest{new("LG1"), new("pw1")},
		},
		{
			"Логин написан в смешанном регистре", loginRequest{new("ITs_mE"), new("mega_password")},
		},
		{
			"Логин написан с пробелами", loginRequest{new("    \tlg2  "), new("pw2")},
		},
		{
			"Email в точности присутствует в бд", loginRequest{new("email3@test.ru"), new("pw3")},
		},
		{
			"Смешанный email", loginRequest{new("     its_me@yandex.ru"), new("mega_password")},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(h.Login))

	defer ts.Close()

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.loginRequest)
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

func TestLoginNegativeBadRequest(t *testing.T) {
	db := storage.NewInMemoryDB()
	h := NewAuthHandler([]byte("SECRET"), "TEST", db)
	usersTearUp(db)

	testcases := []struct {
		name string
		body any
	}{
		{
			name: "Невалидный JSON",
			body: "{loginOrEmail: 'lg1', password:}",
		},
		{
			name: "Отсутствует обязательное поле loginOrEmail",
			body: map[string]string{
				"password": "pw1",
			},
		},
		{
			name: "Пустое тело запроса",
			body: "",
		},
		{
			name: "Рандомный JSON",
			body: "{loginOrEmail: 'l1321', chtoto: 'lads', rand: 12}",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(h.Login))
	defer ts.Close()

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var payload []byte
			var err error

			switch v := tt.body.(type) {
			case string:
				payload = []byte(v)
			default:
				payload, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("ошибка сериализации JSON: %v", err)
				}
			}

			res, err := http.Post(
				ts.URL+loginURL,
				"application/json",
				bytes.NewBuffer(payload),
			)
			if err != nil {
				t.Fatalf("ошибка выполнения запроса: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("ожидался статус 400, получен %d", res.StatusCode)
			}
		})
	}
}

func TestLoginNegativeUnauthorized(t *testing.T) {
	db := storage.NewInMemoryDB()
	h := NewAuthHandler([]byte("SECRET"), "TEST", db)
	usersTearUp(db)

	testcases := []struct {
		name string
		loginRequest
	}{
		{
			"Логин не существует", loginRequest{new("asd"), new("s")},
		},
		{
			"Email не существует", loginRequest{new("x@email.ru"), new("pas")},
		},
		{
			"Неверный пароль", loginRequest{new("lg2"), new("WRONG_PASSWORD_PLS")},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(h.Login))
	defer ts.Close()

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.loginRequest)
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

			if res.StatusCode != http.StatusUnauthorized {
				t.Fatalf("ожидался статус 401, получен %d", res.StatusCode)
			}
		})
	}
}
