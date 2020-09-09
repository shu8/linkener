package stores

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"
	"url-shortener/internal/config"
)

// JSONStore - simple Store based on a JSON file
type JSONStore struct{}

func openFile(write bool) (*os.File, error) {
	var fileNeedsInit = false
	var err error

	if _, err := os.Stat(config.Config.StoreLocation); os.IsNotExist(err) {
		fileNeedsInit = true
		write = true
	}

	var flag int
	if write {
		flag = os.O_CREATE | os.O_RDWR
	} else {
		flag = os.O_RDONLY | os.O_CREATE
	}

	file, err := os.OpenFile(config.Config.StoreLocation, flag, os.ModePerm)
	if err != nil {
		println(err.Error())
		err = errors.New("Failed to open URLs JSON file")
	}

	if fileNeedsInit {
		if _, err := file.Write([]byte("[]")); err != nil {
			println(err.Error())
			err = errors.New("Failed to create URLs JSON file")
		}

		if err := file.Sync(); err != nil {
			println(err.Error())
			err = errors.New("Failed to create URLs JSON file")
		}

		if _, err := file.Seek(0, 0); err != nil {
			println(err.Error())
			err = errors.New("Failed to create URLs JSON file")
		}
	}

	return file, err
}

// GetURLs - GET requests
func (e JSONStore) GetURLs() (*[]ShortURL, error) {
	file, err := openFile(false)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	_, err = decoder.Token()
	if err != nil {
		return nil, errors.New("Failed to parse URLs JSON file: invalid JSON")
	}

	urls := []ShortURL{}
	url := ShortURL{}
	for decoder.More() {
		err := decoder.Decode(&url)
		if err != nil {
			return nil, errors.New("Failed to parse URLs JSON file: invalid JSON")
		}
		urls = append(urls, url)
	}

	return &urls, nil
}

// GetURL - GET /slug requests
func (e JSONStore) GetURL(slug string) (*ShortURL, error) {
	file, err := openFile(false)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	_, err = decoder.Token()
	if err != nil {
		println(err.Error())
		return nil, errors.New("Failed to parse URLs JSON file: invalid JSON")
	}

	url := ShortURL{}
	for decoder.More() {
		err := decoder.Decode(&url)
		if err != nil {
			println(err.Error())
			return nil, errors.New("Failed to parse URLs JSON file: invalid JSON")
		}
		if url.Slug == slug {
			return &url, nil
		}
	}

	return nil, nil
}

// InsertURL - POST requests
func (e JSONStore) InsertURL(slug, url, password string, allowedVisits int) (*ShortURL, error) {
	file, err := openFile(true)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.New("Failed to read URLs JSON file")
	}

	urls := []ShortURL{}
	parseErr := json.Unmarshal(byteValue, &urls)
	if parseErr != nil {
		return nil, errors.New("Failed to parse URLs JSON file")
	}

	newURL := ShortURL{DateCreated: time.Now(), URL: url, Slug: slug, AllowedVisits: allowedVisits, Password: password}
	urls = append(urls, newURL)

	out, err := json.MarshalIndent(urls, "", "    ")
	if err != nil {
		return nil, errors.New("Error saving new URLs JSON file")
	}

	_, err = file.WriteAt(out, 0)
	if err != nil {
		return nil, errors.New("Error writing to URLs JSON file")
	}

	return &newURL, nil
}

// DeleteURL - DELETE requests
func (e JSONStore) DeleteURL(slug string) error {
	file, err := openFile(true)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	_, err = decoder.Token()
	if err != nil {
		return errors.New("Failed to parse URLs JSON file: invalid JSON")
	}

	urls := []ShortURL{}
	parsedURL := ShortURL{}
	found := false
	for decoder.More() {
		err := decoder.Decode(&parsedURL)
		if err != nil {
			return errors.New("Failed to parse URLs JSON file: invalid JSON")
		}

		if parsedURL.Slug == slug {
			found = true
		}

		if parsedURL.Slug != slug {
			urls = append(urls, parsedURL)
		}
	}

	if !found {
		return errors.New("URL not found")
	}

	out, err := json.MarshalIndent(urls, "", "    ")
	if err != nil {
		return errors.New("Error saving new URLs JSON file")
	}

	file.Truncate(0)
	_, err = file.WriteAt(out, 0)
	if err != nil {
		return errors.New("Error writing to URLs JSON file")
	}

	return nil
}

// UpdateURL - PUT requests
func (e JSONStore) UpdateURL(slug, url, password string, allowedVisits int) error {
	file, err := openFile(true)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	_, err = decoder.Token()
	if err != nil {
		return errors.New("Failed to parse URLs JSON file: invalid JSON")
	}

	found := false
	parsedURL := ShortURL{}
	urls := []ShortURL{}

	for decoder.More() {
		err := decoder.Decode(&parsedURL)
		if err != nil {
			return errors.New("Failed to parse URLs JSON file: invalid JSON")
		}
		if parsedURL.Slug == slug {
			(&parsedURL).URL = url
			(&parsedURL).AllowedVisits = allowedVisits
			(&parsedURL).Password = password
			found = true
		}

		urls = append(urls, parsedURL)
	}

	if !found {
		return errors.New("URL not found")
	}

	out, err := json.MarshalIndent(urls, "", "    ")
	if err != nil {
		return errors.New("Error saving new URLs JSON file")
	}

	file.Truncate(0)
	_, err = file.WriteAt(out, 0)
	if err != nil {
		return errors.New("Error writing to URLs JSON file")
	}

	return nil
}

// RecordVisit - record a visit to a short URL
func (e JSONStore) RecordVisit(slug string) error {
	// TODO add more stats like referrer
	file, err := openFile(true)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	_, err = decoder.Token()
	if err != nil {
		return errors.New("Failed to parse URLs JSON file: invalid JSON")
	}

	found := false
	parsedURL := ShortURL{}
	urls := []ShortURL{}

	for decoder.More() {
		err := decoder.Decode(&parsedURL)
		if err != nil {
			return errors.New("Failed to parse URLs JSON file: invalid JSON")
		}
		if parsedURL.Slug == slug {
			(&parsedURL).Stats.AddVisit()
			found = true
		}

		urls = append(urls, parsedURL)
	}

	if !found {
		return errors.New("URL not found")
	}

	out, err := json.MarshalIndent(urls, "", "    ")
	if err != nil {
		return errors.New("Error saving new URLs JSON file")
	}

	file.Truncate(0)
	_, err = file.WriteAt(out, 0)
	if err != nil {
		return errors.New("Error writing to URLs JSON file")
	}

	return nil
}
