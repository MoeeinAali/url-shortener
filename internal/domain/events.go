package domain

type LinkCreatedEvent struct {
	ShortCode string `json:"short_code"`
	LongURL   string `json:"long_url"`
}
type LinkClickedEvent struct {
	ShortCode string `json:"short_code"`
}

type LinkDisabledEvent struct {
	ShortCode string `json:"short_code"`
}
