package stores

import (
	"database/sql"
	"errors"
	"time"
	"github.com/shu8/linkener/internal/config"
)

// SQLiteStore - simple Store for an SQLite database
type SQLiteStore struct{}

var schema = []string{
	`CREATE TABLE IF NOT EXISTS urls (
		slug TEXT PRIMARY KEY,
		url TEXT,
		date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
		allowed_visits INT,
		password TEXT
	);`,
	`CREATE TABLE IF NOT EXISTS url_visits (
		id INT PRIMARY KEY,
		slug TEXT,
		referer TEXT
	);`,
}

func openDB() (*sql.DB, error) {
	// go-sqlite3 will create db if it doesn't exist
	db, err := sql.Open("sqlite3", config.Config.SQLiteStoreLocation)
	if err != nil {
		println(err.Error())
		return nil, errors.New("Unable to open database at " + config.Config.SQLiteStoreLocation)
	}

	if err != nil {
		println(err.Error())
		return nil, errors.New("Unable to initialise database")
	}

	for _, query := range schema {
		_, err := db.Exec(query)
		if err != nil {
			println(err.Error())
			return nil, errors.New("Unable to initialise database")
		}
	}

	return db, nil
}

func getVisits(db *sql.DB, url *ShortURL) error {
	url.Visits = []Visit{}
	rows, err := db.Query("SELECT referer FROM url_visits WHERE slug=?", url.Slug)
	if err != nil {
		println(err.Error())
		return errors.New("Error reading from database")
	}
	defer rows.Close()

	for rows.Next() {
		var referer string
		err := rows.Scan(&referer)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			println(err.Error())
			return errors.New("Error reading from database")
		}
		url.AddVisit(referer)
	}

	return nil
}

// GetURLs - GET requests
func (e SQLiteStore) GetURLs() (*[]ShortURL, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT slug, url, date_created, allowed_visits, password FROM urls")
	if err != nil {
		println(err.Error())
		return nil, errors.New("Error reading from database")
	}
	defer rows.Close()

	urls := []ShortURL{}
	for rows.Next() {
		url := ShortURL{}
		err := rows.Scan(&url.Slug, &url.URL, &url.DateCreated, &url.AllowedVisits, &url.Password)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			println(err.Error())
			return nil, errors.New("Error reading from database")
		}

		err = getVisits(db, &url)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	return &urls, nil
}

// GetURL - GET /slug requests
func (e SQLiteStore) GetURL(slug string) (*ShortURL, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT slug, url, date_created, allowed_visits, password FROM urls WHERE slug=?", slug)

	url := ShortURL{}
	err = row.Scan(&url.Slug, &url.URL, &url.DateCreated, &url.AllowedVisits, &url.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		println(err.Error())
		return nil, errors.New("Error reading from database")
	}

	err = getVisits(db, &url)
	if err != nil {
		return nil, err
	}

	return &url, nil
}

// InsertURL - POST requests
func (e SQLiteStore) InsertURL(slug, url, password string, allowedVisits int) (*ShortURL, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	newURL := ShortURL{DateCreated: time.Now(), URL: url, Slug: slug, AllowedVisits: allowedVisits, Password: password}
	_, err = db.Exec("INSERT INTO urls (slug, url, password, allowed_visits) VALUES(?,?,?,?)", slug, url, password, allowedVisits)
	if err != nil {
		println(err.Error())
		return nil, errors.New("Error saving to database")
	}

	return &newURL, nil
}

// DeleteURL - DELETE requests
func (e SQLiteStore) DeleteURL(slug string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		println(err.Error())
		return errors.New("Error writing to database")
	}

	_, err = tx.Exec("DELETE FROM urls WHERE slug=?", slug)
	if err != nil {
		println(err.Error())
		tx.Rollback()
		return errors.New("Error writing to database")
	}

	_, err = tx.Exec("DELETE FROM url_visits WHERE slug=?", slug)
	if err != nil {
		println(err.Error())
		tx.Rollback()
		return errors.New("Error writing to database")
	}

	err = tx.Commit()
	if err != nil {
		println(err.Error())
		return errors.New("Error writing to database")
	}

	return nil
}

// UpdateURL - PUT requests
func (e SQLiteStore) UpdateURL(slug, url, password string, allowedVisits int) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE urls SET url=?, password=? WHERE slug=?", slug)
	if err != nil {
		println(err.Error())
		return errors.New("Error writing to database")
	}

	return nil
}

// RecordVisit - record a visit to a short URL
func (e SQLiteStore) RecordVisit(slug, referer string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO url_visits (slug, referer) VALUES (?, ?)", slug, referer)
	if err != nil {
		println(err.Error())
		return errors.New("Error writing to database")
	}

	return nil
}
