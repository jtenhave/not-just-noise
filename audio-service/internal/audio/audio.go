package audio

import "time"

type Audio struct {
	ID        string
	CreatorID string
	Title     string
	FileURL   string
	Version   int64
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UpdateAudio struct {
	ID      string
	Title   *string
	FileURL *string
}

type AudioNotifierEventType string

const (
	AudioCreatedEvent AudioNotifierEventType = "CREATED"
	AudioUpdatedEvent AudioNotifierEventType = "UPDATED"
	AudioDeletedEvent AudioNotifierEventType = "DELETED"
)
