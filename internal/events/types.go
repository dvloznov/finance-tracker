package events

import "time"

// DocumentUploadedEventType is emitted when a document has been uploaded and stored.
const DocumentUploadedEventType = "document.uploaded.v1"

// DocumentUploadedEvent represents a document upload notification.
type DocumentUploadedEvent struct {
	EventID    string    `json:"event_id"`
	Type       string    `json:"type"`
	DocumentID string    `json:"document_id"`
	GCSURI     string    `json:"gcs_uri"`
	Filename   string    `json:"filename"`
	UploadedAt time.Time `json:"uploaded_at"`
	Source     string    `json:"source"`
}
