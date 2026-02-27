package link

type LinkCreateRequest struct {
	Url string `json:"url" validate:"required,url"`
}

type LinkUpdateResponse struct {
	Url  string `json:"url" validate:"required,url"`
	Hash string `json:"hash"`
}
