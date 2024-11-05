package modelresponses

type GetUrlStatisticShortenResponse struct {
	Id          int    `json:"id"`
	Url         string `json:"url"`
	ShortCode   string `json:"shortCode"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	AccessCount int    `json:"accessCount"`
}
