package stores

import "time"

// Store - interface for all types of URL data storage formats (e.g. JSON/PostgreSQL)
type Store interface {
	GetURLs() (*[]ShortURL, error)
	GetURL(string) (*ShortURL, error)
	InsertURL(slug, url, password string, allowedVisits int) (*ShortURL, error)
	DeleteURL(slug string) error
	UpdateURL(slug, url, password string, allowedVisits int) error
	RecordVisit(slug, referer string) error
}

// Visit - global structure for each ShortURL
type Visit struct {
	Referer string
}

// AddVisit - helper function to record a visit to a short URL
func (e *ShortURL) AddVisit(referer string) {
	visit := Visit{Referer: referer}
	(*e).Visits = append(e.Visits, visit)
}

// ShortURL - global structure (no matter what Store interface!)
type ShortURL struct {
	Slug          string    `json:"slug"`
	URL           string    `json:"url"`
	DateCreated   time.Time `json:"date_created"`
	AllowedVisits int       `json:"allowed_visits"`
	Visits        []Visit   `json:"visits"`
	// Hide password from JSON responses
	Password string `json:"-"`
}
