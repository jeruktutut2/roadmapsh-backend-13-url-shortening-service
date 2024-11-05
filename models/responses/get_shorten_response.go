package modelresponses

type GetShortenResponse struct {
	Id        int    `json:"id"`
	Url       string `json:"url"`
	ShortCode string `json:"shortCode"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
