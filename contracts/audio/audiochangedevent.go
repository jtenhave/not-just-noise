package audio

type AudioChangedEventType string

const (
	AudioChangedEventTypeCreated AudioChangedEventType = "CREATED"
	AudioChangedEventTypeUpdated AudioChangedEventType = "UPDATED"
	AudioChangedEventTypeDeleted AudioChangedEventType = "DELETED"
)

type AudioChangedEvent struct {
	ID          string                `json:"id"`
	Title       *string               `json:"title"`
	FileURL     *string               `json:"file_url"`
	FileChanged bool                  `json:"file_changed"`
	Version     int64                 `json:"version"`
	Status      string                `json:"status"`
	EventType   AudioChangedEventType `json:"event_type"`
}
