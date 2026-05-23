package notify

import (
	"encoding/json"
	"fmt"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
)

type audioChangedMessage struct {
	EventType audio.AudioNotifierEventType `json:"event_type"`
	AudioID   string                       `json:"id"`
	Title     string                       `json:"title"`
	FileURL   string                       `json:"file_url"`
	Version   int64                        `json:"version"`
}

type AudioNotifyFormatter interface {
	NotifyAudioChangedPayload(eventType audio.AudioNotifierEventType, audio audio.Audio) (string, error)
}

type audioFormatter struct {}

func NewAudioNotifyFormatter() AudioNotifyFormatter {
	return &audioFormatter{}
}

func (formatter audioFormatter) NotifyAudioChangedPayload(eventType audio.AudioNotifierEventType, audio audio.Audio) (string, error) {
	message := audioChangedMessage{
		EventType: eventType,
		AudioID:   audio.ID,
		Title:     audio.Title,
		FileURL:   audio.FileURL,
		Version:   audio.Version,
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("audionotifier.NotifyAudioChangedPayload: failed to marshal message: %w", err)
	}

	return string(jsonMessage), nil
}
