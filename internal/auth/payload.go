package auth

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Is2FAEnabled bool   `json:"is2faEnabled"`
	Email        string `json:"email"`
	Token        string `json:"token"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"Password"`
	Name     string `json:"name" validate:"required"`
}

type RegisterResponse struct {
	Email string `json:"email"`
	Token string `json:"token"`
}
