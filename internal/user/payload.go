package user

type VerifyRequest struct {
	Code  string `json:"code"`
	Email string `json:"email"`
}

type VerifyResponse struct {
	Token string `json:"token"`
}
