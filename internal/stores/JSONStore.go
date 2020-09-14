package stores

import (
	"encoding/json"
	"errors"
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

func writeURLsToFile(file *os.File, urls []ShortURL) error {
	out, err := json.MarshalIndent(urls, "", "    ")
	if err != nil {
		println(err.Error())
		return errors.New("Error saving new URLs JSON file")
	}

	file.Truncate(0)
	_, err = file.WriteAt(out, 0)
	if err != nil {
		println(err.Error())
		return errors.New("Error writing to URLs JSON file")
	}

	return nil
}

func getFileAndDecoder(write bool) (*os.File, *json.Decoder, error) {
	file, err := openFile(write)
	if err != nil {
		return nil, nil, err
	}

	decoder := json.NewDecoder(file)
	_, err = decoder.Token()
	if err != nil {
		file.Close()
		println(err.Error())
		return nil, nil, errors.New("Failed to parse URLs JSON file: invalid JSON")
	}

	return file, decoder, nil
}

func getAllURLs(decoder *json.Decoder) ([]ShortURL, error) {
	urls := []ShortURL{}
	for decoder.More() {
		parsedURL := ShortURL{}
		err := decoder.Decode(&parsedURL)
		if err != nil {
			println(err.Error())
			return nil, errors.New("Failed to parse URLs JSON file: invalid JSON")
		}
		urls = append(urls, parsedURL)
	}

	return urls, nil
}

// GetURLs - GET requests
func (e JSONStore) GetURLs() (*[]ShortURL, error) {
	file, decoder, err := getFileAndDecoder(false)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	urls, err := getAllURLs(decoder)
	if err != nil {
		return nil, err
	}

	return &urls, nil
}

// GetURL - GET /slug requests
func (e JSONStore) GetURL(slug string) (*ShortURL, error) {
	file, decoder, err := getFileAndDecoder(false)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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
	file, decoder, err := getFileAndDecoder(true)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	urls, err := getAllURLs(decoder)
	if err != nil {
		return nil, err
	}

	newURL := ShortURL{DateCreated: time.Now(), URL: url, Slug: slug, AllowedVisits: allowedVisits, Password: password}
	urls = append(urls, newURL)

	if err := writeURLsToFile(file, urls); err != nil {
		return nil, err
	}

	return &newURL, nil
}

// DeleteURL - DELETE requests
func (e JSONStore) DeleteURL(slug string) error {
	file, decoder, err := getFileAndDecoder(true)
	if err != nil {
		return err
	}
	defer file.Close()

	urls := []ShortURL{}
	found := false
	for decoder.More() {
		parsedURL := ShortURL{}
		err := decoder.Decode(&parsedURL)
		if err != nil {
			println(err.Error())
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

	if err := writeURLsToFile(file, urls); err != nil {
		return err
	}

	return nil
}

// UpdateURL - PUT requests
func (e JSONStore) UpdateURL(slug, url, password string, allowedVisits int) error {
	file, decoder, err := getFileAndDecoder(true)
	if err != nil {
		return err
	}
	defer file.Close()

	found := false
	urls := []ShortURL{}

	for decoder.More() {
		parsedURL := ShortURL{}
		err := decoder.Decode(&parsedURL)
		if err != nil {
			println(err.Error())
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

	if err := writeURLsToFile(file, urls); err != nil {
		return err
	}

	return nil
}

// RecordVisit - record a visit to a short URL
func (e JSONStore) RecordVisit(slug, referer string) error {
	file, decoder, err := getFileAndDecoder(true)
	if err != nil {
		return err
	}
	defer file.Close()

	found := false
	urls := []ShortURL{}

	for decoder.More() {
		parsedURL := ShortURL{}
		err := decoder.Decode(&parsedURL)
		if err != nil {
			println(err.Error())
			return errors.New("Failed to parse URLs JSON file: invalid JSON")
		}
		if parsedURL.Slug == slug {
			(&parsedURL).AddVisit(referer)
			found = true
		}

		urls = append(urls, parsedURL)
	}

	if !found {
		return errors.New("URL not found")
	}

	if err := writeURLsToFile(file, urls); err != nil {
		return err
	}

	return nil
}
