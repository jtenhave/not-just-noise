package audio

import "time"

type Audio struct {
	ID        string
	Title     string
	Creator   string
	FileURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
