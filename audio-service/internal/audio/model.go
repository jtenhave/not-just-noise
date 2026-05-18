package audio

import "time"

type Audio struct {
	ID        string
	CreatorID string
	Title     string
	FileURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
