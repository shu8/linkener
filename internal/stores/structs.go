package stores

import "time"

// Store - interface for all types of URL data storage formats (e.g. JSON/PostgreSQL)
type Store interface {
	GetURLs() (*[]ShortURL, error)
	GetURL(string) (*ShortURL, error)
	InsertURL(slug, url, password string, allowedVisits int) (*ShortURL, error)
	DeleteURL(slug string) error
	UpdateURL(slug, url, password string, allowedVisits int) error
	RecordVisit(slug string) error
}

// VisitStats - global structure for each ShortURL
type VisitStats struct {
	Count int `json:"count"`
}

// AddVisit - helper function to record a visit to a short URL
func (e *VisitStats) AddVisit() {
	(*e).Count++
}

// ShortURL - global structure (no matter what Store interface!)
type ShortURL struct {
	Slug          string     `json:"slug"`
	URL           string     `json:"url"`
	DateCreated   time.Time  `json:"date_created"`
	AllowedVisits int        `json:"allowed_visits"`
	Stats         VisitStats `json:"stats"`
	Password      string     `json:"password"`
}
