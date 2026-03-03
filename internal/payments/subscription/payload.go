package subscription

type SetupIntentResponse struct {
	ClientSecret string `json:"clientSecret"`
}

type CreateSubscriptionPayload struct {
	PlanId uint `json:"planId" validate:"required"`
}
