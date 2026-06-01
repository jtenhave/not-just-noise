package audio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	njnHttp "github.com/jtenhave/not-just-noise/lib/http"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type AudioService interface {
	GetAudio(context context.Context, id string) (Audio, error)
	CreateAudio(context context.Context, audio Audio) (string, error)
	UpdateAudio(context context.Context, audio Audio) error
	DeleteAudio(context context.Context, id string) error
}

const (
	idIsRequired      = "id is required"
	titleIsRequired   = "title is required"
	creatorIsRequired = "creator is required"
	fileURLIsRequired = "file_url is required"
	fileURLIsNotValid = "file_url is not a valid URL"
	failedToUnmarshal = "failed to unmarshal"
)

type ErrorResponseBody struct {
	Error string `json:"error"`
}

// CreateRoutes creates the routes using the given audioService.
func CreateRoutes(audioService AudioService) []njnHttp.Route {
	return []njnHttp.Route{

		{
			Method: "GET",
			Path:   "/audio/{id}",
			Handler: func(request *http.Request, responseWriter http.ResponseWriter) {
				getAudioHandler(request, responseWriter, audioService)
			},
		},
		{
			Method: "POST",
			Path:   "/audio",
			Handler: func(request *http.Request, responseWriter http.ResponseWriter) {
				createAudioHandler(request, responseWriter, audioService)
			},
		},

		{
			Method: "PATCH",
			Path:   "/audio/{id}",
			Handler: func(request *http.Request, responseWriter http.ResponseWriter) {
				updateAudioHandler(request, responseWriter, audioService)
			},
		},
		{
			Method: "DELETE",
			Path:   "/audio/{id}",
			Handler: func(request *http.Request, responseWriter http.ResponseWriter) {
				deleteAudioHandler(request, responseWriter, audioService)
			},
		},
	}
}

type AudioResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Creator   string    `json:"creator"`
	FileURL   string    `json:"file_url"`
	Version   int64     `json:"version"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToAudioResponse converts the given audio to an AudioResponse.
func toAudioResponse(audio Audio) AudioResponse {
	return AudioResponse{
		ID:        audio.ID,
		Title:     *audio.Title,
		Creator:   audio.CreatorID,
		FileURL:   *audio.FileURL,
		Version:   audio.Version,
		Status:    audio.Status,
		CreatedAt: audio.CreatedAt,
		UpdatedAt: audio.UpdatedAt,
	}
}

// getAudioHandler handles the GET /audio/{id} request, using the given audioService.
func getAudioHandler(request *http.Request, responseWriter http.ResponseWriter, audioService AudioService) {
	id := request.PathValue("id")
	if id == "" {
		njnHttp.SendErrorResponse(responseWriter, njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.getAudioHandler: %s", idIsRequired)))
		return
	}

	audio, err := audioService.GetAudio(request.Context(), id)
	if err != nil {
		njnHttp.SendErrorResponse(responseWriter, njnerror.Wrapf("audiohandler.getAudioHandler: failed to get audio: %w", err))
		return
	}

	njnHttp.SendJsonResponse(responseWriter, 200, toAudioResponse(audio))
}

type CreateAudioRequest struct {
	Title     string `json:"title"`
	CreatorID string `json:"creator_id"`
	FileURL   string `json:"file_url"`
}

type CreateAudioResponse struct {
	ID string `json:"id"`
}

// ToAudio converts the given createAudioRequest to an Audio.
func (createAudioRequest CreateAudioRequest) ToAudio() Audio {
	return Audio{
		Title:     &createAudioRequest.Title,
		CreatorID: createAudioRequest.CreatorID,
		FileURL:   &createAudioRequest.FileURL,
	}
}

// Validate validates the given createAudioRequest.
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

// createAudioHandler handles the POST /audio request, using the given audioService.
func createAudioHandler(request *http.Request, responseWriter http.ResponseWriter, audioService AudioService) {
	createAudioRequest := CreateAudioRequest{}
	err := njnHttp.ReadAllAndUnmarshal(request.Body, &createAudioRequest)
	if err != nil {
		njnHttp.SendErrorResponse(responseWriter, njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.createAudioHandler: %s create audio request", failedToUnmarshal)))
		return
	}

	errors := createAudioRequest.Validate()
	if len(errors) > 0 {
		njnHttp.SendErrorResponse(responseWriter, njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.createAudioHandler: %s", strings.Join(errors, ", "))))
		return
	}

	id, err := audioService.CreateAudio(request.Context(), createAudioRequest.ToAudio())
	if err != nil {
		njnHttp.SendErrorResponse(responseWriter, njnerror.Wrapf("audiohandler.createAudioHandler: failed to create audio: %w", err))
		return
	}

	response := CreateAudioResponse{
		ID: id,
	}

	njnHttp.SendJsonResponse(responseWriter, 200, response)
}

type UpdateAudioRequest struct {
	Title   *string `json:"title"`
	FileURL *string `json:"file_url"`
}

// Validate validates the given updateAudioRequest.
func (updateAudioRequest UpdateAudioRequest) Validate() []string {
	errors := []string{}
	if updateAudioRequest.Title != nil && *updateAudioRequest.Title == "" {
		errors = append(errors, titleIsRequired)
	}

	if updateAudioRequest.FileURL != nil {
		if *updateAudioRequest.FileURL == "" {
			errors = append(errors, fileURLIsRequired)
		} else if !isValidURL(*updateAudioRequest.FileURL) {
			errors = append(errors, fileURLIsNotValid)
		}
	}
	return errors
}

// ToAudio converts the given updateAudioRequest to an Audio.
func (updateAudioRequest UpdateAudioRequest) ToAudio() Audio {
	return Audio{
		Title:   updateAudioRequest.Title,
		FileURL: updateAudioRequest.FileURL,
	}
}

// updateAudioHandler handles the PATCH /audio/{id} request, using the given audioService.
func updateAudioHandler(request *http.Request, responseWriter http.ResponseWriter, audioService AudioService) {
	id := request.PathValue("id")
	if id == "" {
		njnHttp.SendErrorResponse(responseWriter, njnerror.NewNJNError(njnerror.BadRequest, idIsRequired))
		return
	}

	updateAudioRequest := UpdateAudioRequest{}
	err := njnHttp.ReadAllAndUnmarshal(request.Body, &updateAudioRequest)
	if err != nil {
		njnHttp.SendErrorResponse(responseWriter, njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.updateAudioHandler: %s update audio request", failedToUnmarshal)))
		return
	}

	errors := updateAudioRequest.Validate()
	if len(errors) > 0 {
		njnHttp.SendErrorResponse(responseWriter, njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.updateAudioHandler: %s", strings.Join(errors, ", "))))
		return
	}

	audio := updateAudioRequest.ToAudio()
	audio.ID = id

	err = audioService.UpdateAudio(request.Context(), audio)
	if err != nil {
		njnHttp.SendErrorResponse(responseWriter, njnerror.Wrapf("audiohandler.updateAudioHandler: failed to update audio: %w", err))
		return
	}

	njnHttp.SendJsonResponse(responseWriter, 204, nil)
}

// deleteAudioHandler handles the DELETE /audio/{id} request, using the given audioService.
func deleteAudioHandler(request *http.Request, responseWriter http.ResponseWriter, audioService AudioService) {
	id := request.PathValue("id")
	if id == "" {
		njnHttp.SendErrorResponse(responseWriter, njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.deleteAudioHandler: %s", idIsRequired)))
		return
	}

	err := audioService.DeleteAudio(request.Context(), id)
	if err != nil {
		njnHttp.SendErrorResponse(responseWriter, njnerror.Wrapf("audiohandler.deleteAudioHandler: failed to delete audio: %w", err))
		return
	}

	njnHttp.SendJsonResponse(responseWriter, 204, nil)
}

// isValidURL checks if the given str is a valid URL.
func isValidURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
