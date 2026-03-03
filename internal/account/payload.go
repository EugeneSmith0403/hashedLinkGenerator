package account

type CreateAccountResponse struct {
	ID            uint            `json:"id"`
	AccountStatus AccountStatus   `json:"accountStatus"`
	Provider      PaymentProvider `json:"provider"`
}

type UpdateAccountRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UpdateAccountResponse struct {
	ID            uint            `json:"id"`
	AccountStatus AccountStatus   `json:"accountStatus"`
	Provider      PaymentProvider `json:"provider"`
}
