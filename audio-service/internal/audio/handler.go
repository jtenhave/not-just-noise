package audio

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jtenhave/not-just-noise/lib/errorcode"
	"github.com/jtenhave/not-just-noise/lib/http"
)

type AudioService interface {
	GetAudio(id string) (Audio, error)
	CreateAudio(audio Audio) (string, error)
	UpdateAudio(audio Audio) (Audio, error)
	DeleteAudio(id string) error
}

const (
	failedToUnmarshal = "failed to unmarshal"
	failedToMarshal   = "failed to marshal"
	idIsRequired      = "id is required"
	titleIsRequired   = "title is required"
	creatorIsRequired = "creator is required"
	fileURLIsRequired = "file_url is required"
	fileURLIsNotValid = "file_url is not a valid URL"
	audioNotFound     = "audio not found"
)

func CreateRoutes(audioService AudioService) []http.Route {
	return []http.Route{
		http.CreateRoute("GET", "/audio/{id}", func(request http.Request) http.Response {
			return getAudioHandler(request, audioService)
		}),

		http.CreateRoute("POST", "/audio", func(request http.Request) http.Response {
			return createAudioHandler(request, audioService)
		}),

		http.CreateRoute("PATCH", "/audio/{id}", func(request http.Request) http.Response {
			return updateAudioHandler(request, audioService)
		}),

		http.CreateRoute("DELETE", "/audio/{id}", func(request http.Request) http.Response {
			return deleteAudioHandler(request, audioService)
		}),
	}
}

type AudioResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Creator   string    `json:"creator"`
	FileURL   string    `json:"file_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (audio Audio) ToAudioResponse() AudioResponse {
	return AudioResponse{
		ID:        audio.ID,
		Title:     audio.Title,
		Creator:   audio.CreatorID,
		FileURL:   audio.FileURL,
		CreatedAt: audio.CreatedAt,
		UpdatedAt: audio.UpdatedAt,
	}
}

func getAudioHandler(request http.Request, audioService AudioService) http.Response {
	id := request.PathValue("id")
	if id == "" {
		return http.CreateErrorResonse(400, idIsRequired)
	}

	audio, err := audioService.GetAudio(id)
	if err != nil {
		if errorcode.ErrorCode(err) == errorcode.NotFound {
			return http.CreateErrorResonse(404, audioNotFound)
		}

		return http.CreateErrorResonse(500, err.Error())
	}

	resultJson, err := json.Marshal(audio.ToAudioResponse())
	if err != nil {
		return http.CreateErrorResonse(500, fmt.Errorf("%s get audio response: %w", failedToMarshal, err).Error())
	}

	return http.Response{
		StatusCode: 200,
		Body:       string(resultJson),
	}
}

type CreateAudioRequest struct {
	Title     string `json:"title"`
	CreatorID string `json:"creator_id"`
	FileURL   string `json:"file_url"`
}

func (createAudioRequest CreateAudioRequest) ToAudio() Audio {
	return Audio{
		Title:     createAudioRequest.Title,
		CreatorID: createAudioRequest.CreatorID,
		FileURL:   createAudioRequest.FileURL,
	}
}

func (createAudioRequest CreateAudioRequest) Validate() []string {
	errors := []string{}
	if createAudioRequest.Title == "" {
		errors = append(errors, titleIsRequired)
	}
	if createAudioRequest.CreatorID == "" {
		errors = append(errors, creatorIsRequired)
	}
	if createAudioRequest.FileURL == "" {
		errors = append(errors, fileURLIsRequired)
	} else if !isValidURL(createAudioRequest.FileURL) {
		errors = append(errors, fileURLIsNotValid)
	}
	return errors
}

func createAudioHandler(request http.Request, audioService AudioService) http.Response {
	var createAudioRequest CreateAudioRequest
	err := json.Unmarshal([]byte(request.Body), &createAudioRequest)
	if err != nil {
		return http.CreateErrorResonse(400, fmt.Errorf("%s create audio request: %w", failedToUnmarshal, err).Error())
	}

	errors := createAudioRequest.Validate()
	if len(errors) > 0 {
		return http.CreateErrorResonse(400, strings.Join(errors, ", "))
	}

	id, err := audioService.CreateAudio(createAudioRequest.ToAudio())
	if err != nil {
		code := errorcode.ErrorCode(err)
		if code == 0 {
			code = 500
		}

		return http.CreateErrorResonse(code, err.Error())
	}

	response := map[string]string{
		"id": id,
	}

	resultJson, err := json.Marshal(response)
	if err != nil {
		return http.CreateErrorResonse(500, fmt.Errorf("%s create audio response: %w", failedToMarshal, err).Error())
	}

	return http.Response{
		StatusCode: 200,
		Body:       string(resultJson),
	}
}

type UpdateAudioRequest struct {
	Title   *string `json:"title"`
	Creator *string `json:"creator"`
	FileURL *string `json:"file_url"`
}

func (updateAudioRequest UpdateAudioRequest) Validate() []string {
	errors := []string{}
	if updateAudioRequest.Title != nil && *updateAudioRequest.Title == "" {
		errors = append(errors, "title cannot be empty")
	}

	if updateAudioRequest.Creator != nil && *updateAudioRequest.Creator == "" {
		errors = append(errors, "creator cannot be empty")
	}

	if updateAudioRequest.FileURL != nil {
		if *updateAudioRequest.FileURL == "" {
			errors = append(errors, "file_url cannot be empty")
		} else if !isValidURL(*updateAudioRequest.FileURL) {
			errors = append(errors, "file_url is not a valid URL")
		}
	}
	return errors
}

func (updateAudioRequest UpdateAudioRequest) ToAudio() Audio {
	return Audio{
		Title:     *updateAudioRequest.Title,
		CreatorID: *updateAudioRequest.Creator,
		FileURL:   *updateAudioRequest.FileURL,
	}
}

func updateAudioHandler(request http.Request, audioService AudioService) http.Response {
	id := request.PathValue("id")
	if id == "" {
		return http.CreateErrorResonse(400, "id is required")
	}

	var updateAudioRequest UpdateAudioRequest
	err := json.Unmarshal([]byte(request.Body), &updateAudioRequest)
	if err != nil {
		return http.CreateErrorResonse(400, fmt.Errorf("Failed to unmarshal update audio request: %w", err).Error())
	}

	errors := updateAudioRequest.Validate()
	if len(errors) > 0 {
		return http.CreateErrorResonse(400, strings.Join(errors, ", "))
	}

	audio := updateAudioRequest.ToAudio()
	audio.ID = id
	audio, err = audioService.UpdateAudio(audio)

	if err != nil {
		return http.CreateErrorResonse(500, fmt.Errorf("Failed to update audio: %w", err).Error())
	}

	resultJson, err := json.Marshal(audio.ToAudioResponse())
	if err != nil {
		return http.CreateErrorResonse(500, fmt.Errorf("Failed to marshal update audio response: %w", err).Error())
	}

	return http.Response{
		StatusCode: 200,
		Body:       string(resultJson),
	}
}

func deleteAudioHandler(request http.Request, audioService AudioService) http.Response {
	id := request.PathValue("id")
	if id == "" {
		return http.CreateErrorResonse(400, "id is required")
	}

	err := audioService.DeleteAudio(id)
	if err != nil {
		return http.CreateErrorResonse(500, fmt.Errorf("Failed to delete audio: %w", err).Error())
	}

	return http.Response{
		StatusCode: 200,
		Body:       "Audio deleted successfully",
	}
}

func isValidURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
