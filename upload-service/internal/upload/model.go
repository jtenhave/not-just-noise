package upload

import "time"

type Upload struct {
	ID        string
	AudioID   string
	FileURL   string
	FileHash  string
	Version   int64
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
