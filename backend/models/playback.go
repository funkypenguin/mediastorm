package models

// PlaybackResolution contains the derived streaming details for an NZB selection.
type PlaybackResolution struct {
	QueueID       int64  `json:"queueId"`
	WebDAVPath    string `json:"webdavPath"`
	HealthStatus  string `json:"healthStatus"`
	FileSize      int64  `json:"fileSize,omitempty"`
	SourceNZBPath string `json:"sourceNzbPath,omitempty"`
}
