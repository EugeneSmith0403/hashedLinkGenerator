package main

import (
	"link-generator/configs"
	"link-generator/internal/auth"
	"link-generator/internal/user"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func loadTestConfig() *configs.Config {
	err := godotenv.Load(".env.test")

	if err != nil {
		panic(err.Error())
	}

	return &configs.Config{
		Db: configs.DbConfig{
			Dsn: os.Getenv("DSN"),
		},
		Auth: configs.AuthConfig{
			Secret: os.Getenv("TOKEN"),
		},
	}
}

func initDb() *gorm.DB {
	config := loadTestConfig()
	db, err := gorm.Open(postgres.Open(config.Db.Dsn), &gorm.Config{})

	if err != nil {
		panic(err.Error())
	}

	return db
}

func initData(db *gorm.DB) {
	db.Create(&user.User{
		Email:    "test1@e.com",
		Password: "$2a$10$rPoAoGvsaQZ/tRWf3DZphuUs1LbWIky0XppainCMrISRcDe8FOH0C",
		Name:     "test",
	})
}

func removeData(db *gorm.DB) {
	db.Unscoped().
		Where("email = ?", "test1@e.com").
		Delete(&user.User{})
}

func request(requestData *auth.LoginRequest, t *testing.T) *http.Response {
	config := loadTestConfig()
	// Init
	ts := httptest.NewServer(App(config))

	defer ts.Close()

	data, _ := json.Marshal(requestData)

	res, err := http.Post(ts.URL+"/auth/login", "application/json", bytes.NewReader(data))

	if err != nil {
		t.Fatal(err)
	}

	return res
}

func TestLoginSuccess(t *testing.T) {
	// Prepare
	db := initDb()
	initData(db)

	res := request(&auth.LoginRequest{
		Email:    "test1@e.com",
		Password: "1",
	}, t)

	if res.StatusCode != 200 {
		t.Fatalf("Expected %d, received %d", 200, res.StatusCode)
	}

	var result map[string]any

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result["token"] == "" {
		t.Fatalf("Expected token")
	}

	removeData(db)
}

func TestLoginFail(t *testing.T) {
	// Prepare
	db := initDb()
	initData(db)
	res := request(&auth.LoginRequest{
		Email:    "test1@e.com",
		Password: "2",
	}, t)

	if res.StatusCode != 401 {
		t.Fatalf("Expected %d, received %d", 401, res.StatusCode)
	}

	removeData(db)
}
