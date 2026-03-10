package auth_test

import (
	"link-generator/configs"
	"link-generator/internal/auth"
	internalJWT "link-generator/internal/jwt"
	"link-generator/internal/user"
	"link-generator/pkg/db"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupAuthHandler создает тестовое окружение для тестов авторизации
func setupAuthHandler(t *testing.T) (*auth.AuthHandler, sqlmock.Sqlmock, func()) {
	t.Helper()

	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	gormDb, err := gorm.Open(postgres.New(postgres.Config{Conn: database}), &gorm.Config{})
	if err != nil {
		database.Close()
		t.Fatalf("Failed to open gorm connection: %v", err)
	}

	userRepo := user.NewUserRepository(&db.Db{DB: gormDb})

	config := &configs.Config{
		Auth: configs.AuthConfig{
			Secret: "test-secret-key",
		},
	}

	jwtService := internalJWT.NewJWTService(internalJWT.JwtDeps{
		Secret: config.Auth.Secret,
	})

	authService := auth.NewAuthService(userRepo)

	authHandler := auth.NewAuthHandlerForTest(config, authService, jwtService)

	cleanup := func() {
		database.Close()
	}

	return authHandler, mock, cleanup
}

func TestLoginSuccess(t *testing.T) {
	authHandler, mock, cleanup := setupAuthHandler(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "email", "password", "name"}).
		AddRow(1, time.Now(), time.Now(), nil, auth.TestEmail, auth.TestPasswordHash, "Test User")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email=$1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(auth.TestEmail, 1).
		WillReturnRows(rows)

	loginRequest := auth.LoginRequest{
		Email:    auth.TestEmail,
		Password: auth.TestPassword,
	}
	requestBody, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := authHandler.Login()
	handler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response auth.LoginResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Email != auth.TestEmail {
		t.Errorf("Expected email %s, got %s", auth.TestEmail, response.Email)
	}

	if response.Token == "" {
		t.Error("Expected token to be non-empty")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	authHandler, mock, cleanup := setupAuthHandler(t)
	defer cleanup()

	wrongPassword := "wrongpassword"

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "email", "password", "name"}).
		AddRow(1, time.Now(), time.Now(), nil, auth.TestEmail, auth.TestPasswordHash, "Test User")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email=$1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(auth.TestEmail, 1).
		WillReturnRows(rows)

	loginRequest := auth.LoginRequest{
		Email:    auth.TestEmail,
		Password: wrongPassword,
	}
	requestBody, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := authHandler.Login()
	handler(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}

func TestLoginUserNotFound(t *testing.T) {
	authHandler, mock, cleanup := setupAuthHandler(t)
	defer cleanup()

	nonexistentEmail := "nonexistent@example.com"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email=$1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(nonexistentEmail, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	loginRequest := auth.LoginRequest{
		Email:    nonexistentEmail,
		Password: auth.TestPassword,
	}
	requestBody, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := authHandler.Login()
	handler(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}
