package model

import "time"

type File struct {
	ID           string     `json:"id"`
	Filename     string     `json:"filename"`
	StoragePath  string     `json:"-"`
	Size         int64      `json:"size"`
	MimeType     string     `json:"mimeType"`
	PasswordHash *string    `json:"-"`
	CreatedAt    time.Time  `json:"createdAt"`
	ExpiresAt    *time.Time `json:"expiresAt"`
	DownloadCount int       `json:"downloadCount"`
}

type FileMeta struct {
	ID             string `json:"id"`
	Filename       string `json:"filename"`
	Size           int64  `json:"size"`
	MimeType       string `json:"mimeType"`
	HasPassword    bool   `json:"hasPassword"`
	CreatedAt      string `json:"createdAt"`
	ExpiresAt      string `json:"expiresAt,omitempty"`
	DownloadCount  int    `json:"downloadCount"`
}

type UploadResponse struct {
	ID       string `json:"id"`
	ShareURL string `json:"shareUrl"`
}

// ConfigResponse tells the frontend the server's upload limits.
type ConfigResponse struct {
	MaxUploadBytes int64 `json:"maxUploadBytes"`
	MaxUploadMB    int64 `json:"maxUploadMB"`
	ExpirySeconds  int64 `json:"expirySeconds"` // 0 = files never expire
}