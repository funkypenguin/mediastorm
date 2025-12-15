package models

// NZBHealthCheck describes the result of checking an NZB against a Usenet server.
type NZBHealthCheck struct {
	Status          string   `json:"status"`
	Healthy         bool     `json:"healthy"`
	CheckedSegments int      `json:"checkedSegments"`
	TotalSegments   int      `json:"totalSegments"`
	MissingSegments []string `json:"missingSegments,omitempty"`
	FileName        string   `json:"fileName,omitempty"`
	Sampled         bool     `json:"sampled,omitempty"`
}
