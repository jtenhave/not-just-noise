package contracts

type UploadStartedMessage struct {
	UploadID string `json:"upload_id"`
	FileURL  string `json:"file_url"`
	FileHash string `json:"file_hash"`
	Version  int64  `json:"version"`
}
