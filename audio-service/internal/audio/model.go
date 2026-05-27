package audio

import "time"

type Audio struct {
	ID        string
	CreatorID string
	Title     *string
	FileURL   *string
	Version   int64
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AudioPublishEvent struct {
	Audio     Audio
	EventType string
}
