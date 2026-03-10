package auth_test

import (
	"link-generator/internal/auth"
	"link-generator/internal/user"
	"testing"
)

type MockUserRepository struct{}

func (repo *MockUserRepository) Create(u *user.User) (*user.User, error) {
	return u, nil
}

func (repo *MockUserRepository) FindByEmail(email string) (*user.User, error) {
	return nil, nil
}

func TestRegisterSucces(t *testing.T) {
	authService := auth.NewAuthService(&MockUserRepository{})
	email, err := authService.Register(auth.TestName, auth.TestEmail, auth.TestPassword)

	if err != nil {
		t.Fatal(err)
	}

	if email != auth.TestEmail {
		t.Fatalf("Expect email")
	}

}
