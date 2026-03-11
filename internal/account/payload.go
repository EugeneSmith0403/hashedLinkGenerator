package account

import "link-generator/internal/models"

type CreateAccountResponse struct {
	ID            uint                  `json:"id"`
	AccountStatus models.AccountStatus  `json:"accountStatus"`
	Provider      models.PaymentProvider `json:"provider"`
}

type UpdateAccountRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UpdateAccountResponse struct {
	ID            uint                  `json:"id"`
	AccountStatus models.AccountStatus  `json:"accountStatus"`
	Provider      models.PaymentProvider `json:"provider"`
}
