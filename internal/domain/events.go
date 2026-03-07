package domain

type LinkCreatedEvent struct {
	ShortCode string `json:"short_code"`
}
type LinkClickedEvent struct {
	ShortCode string `json:"short_code"`
}
