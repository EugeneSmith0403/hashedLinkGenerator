package auth_test

import (
	"link-generator/internal/auth"
	"link-generator/internal/models"
	"testing"
)

type MockUserRepository struct{}

func (repo *MockUserRepository) Create(u *models.User) (*models.User, error) {
	return u, nil
}

func (repo *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	return nil, nil
}

func TestRegisterSucces(t *testing.T) {
	authService := auth.NewAuthService(auth.AuthServiceDeps{UserRepository: &MockUserRepository{}})
	email, err := authService.Register(auth.TestName, auth.TestEmail, auth.TestPassword)

	if err != nil {
		t.Fatal(err)
	}

	if email != auth.TestEmail {
		t.Fatalf("Expect email")
	}

}
