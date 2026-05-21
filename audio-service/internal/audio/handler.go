package audio

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/jtenhave/not-just-noise/lib/http"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type AudioService interface {
	GetAudio(id string) (Audio, error)
	CreateAudio(audio Audio) (string, error)
	UpdateAudio(audio UpdateAudio) error
	DeleteAudio(id string) error
}

const (
	idIsRequired      = "id is required"
	titleIsRequired   = "title is required"
	creatorIsRequired = "creator is required"
	fileURLIsRequired = "file_url is required"
	fileURLIsNotValid = "file_url is not a valid URL"
	audioNotFound     = "audio not found"
	failedToUnmarshal = "failed to unmarshal"
)

func CreateRoutes(audioService AudioService) []http.Route {
	return []http.Route{

		http.CreateRoute("GET", "/audio/{id}", nil, func(request http.Request) http.Response {
			return getAudioHandler(request, audioService)
		}),

		http.CreateRoute("POST", "/audio", reflect.TypeOf(CreateAudioRequest{}), func(request http.Request) http.Response {
			return createAudioHandler(request, audioService)
		}),

		http.CreateRoute("PATCH", "/audio/{id}", reflect.TypeOf(UpdateAudioRequest{}), func(request http.Request) http.Response {
			return updateAudioHandler(request, audioService)
		}),

		http.CreateRoute("DELETE", "/audio/{id}", nil, func(request http.Request) http.Response {
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
	id := request.PathValues["id"]
	if id == "" {
		return http.CreateErrorResponse(njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.getAudioHandler: %s", idIsRequired)))
	}

	audio, err := audioService.GetAudio(id)
	if err != nil {
		return http.CreateErrorResponse(njnerror.Wrapf("audiohandler.getAudioHandler: failed to get audio: %w", err))
	}

	return http.CreateResponse(200, audio.ToAudioResponse())
}

type CreateAudioRequest struct {
	Title     string `json:"title"`
	CreatorID string `json:"creator_id"`
	FileURL   string `json:"file_url"`
}

type CreateAudioResponse struct {
	ID string `json:"id"`
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
	createAudioRequest, ok := request.Body.(CreateAudioRequest)
	if !ok {
		return http.CreateErrorResponse(njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.createAudioHandler: %s create audio request", failedToUnmarshal)))
	}

	errors := createAudioRequest.Validate()
	if len(errors) > 0 {
		return http.CreateErrorResponse(njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.createAudioHandler: %s", strings.Join(errors, ", "))))
	}

	id, err := audioService.CreateAudio(createAudioRequest.ToAudio())
	if err != nil {
		return http.CreateErrorResponse(njnerror.Wrapf("audiohandler.createAudioHandler: failed to create audio: %w", err))
	}

	response := CreateAudioResponse{
		ID: id,
	}

	return http.CreateResponse(200, response)
}

type UpdateAudioRequest struct {
	Title   *string `json:"title"`
	FileURL *string `json:"file_url"`
}

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

func (updateAudioRequest UpdateAudioRequest) ToPatchAudio() UpdateAudio {
	return UpdateAudio{
		Title:   updateAudioRequest.Title,
		FileURL: updateAudioRequest.FileURL,
	}
}

func updateAudioHandler(request http.Request, audioService AudioService) http.Response {
	id := request.PathValues["id"]
	if id == "" {
		return http.CreateErrorResponse(njnerror.NewNJNError(njnerror.BadRequest, idIsRequired))
	}

	updateAudioRequest, ok := request.Body.(UpdateAudioRequest)
	if !ok {
		return http.CreateErrorResponse(njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.updateAudioHandler: %s update audio request", failedToUnmarshal)))
	}

	errors := updateAudioRequest.Validate()
	if len(errors) > 0 {
		return http.CreateErrorResponse(njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.updateAudioHandler: %s", strings.Join(errors, ", "))))
	}

	audio := updateAudioRequest.ToPatchAudio()
	audio.ID = id

	err := audioService.UpdateAudio(audio)
	if err != nil {
		return http.CreateErrorResponse(njnerror.Wrapf("audiohandler.updateAudioHandler: failed to update audio: %w", err))
	}

	return http.CreateResponse(204, nil)
}

func deleteAudioHandler(request http.Request, audioService AudioService) http.Response {
	id := request.PathValues["id"]
	if id == "" {
		return http.CreateErrorResponse(njnerror.NewNJNError(njnerror.BadRequest, fmt.Sprintf("audiohandler.deleteAudioHandler: %s", idIsRequired)))
	}

	err := audioService.DeleteAudio(id)
	if err != nil {
		return http.CreateErrorResponse(njnerror.Wrapf("audiohandler.deleteAudioHandler: failed to delete audio: %w", err))
	}

	return http.CreateResponse(204, nil)
}

func isValidURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
