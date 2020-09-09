package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"url-shortener/internal/config"
	"url-shortener/internal/stores"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type newURLRequest struct {
	URL           string `json:"url"`
	Slug          string `json:"slug"`
	SlugLength    int    `json:"slug_length"`
	AllowedVisits int    `json:"allowed_visits"`
	Password      string `json:"password"`
}

type updateURLRequest struct {
	URL           string `json:"url"`
	AllowedVisits int    `json:"allowed_visits"`
	Password      string `json:"password"`
}

func generateSlug(slugLength int) (string, error) {
	bytes := make([]byte, slugLength*2)

	_, err := rand.Read(bytes[:])
	if err != nil {
		return "", err
	}

	// Don't use / characters, so more URL friendly
	r := strings.NewReplacer("/", "_")
	b64 := base64.StdEncoding.EncodeToString(bytes[:])[:slugLength]

	slug := r.Replace(b64)
	return slug, nil
}

func urlsHandler(w http.ResponseWriter, r *http.Request, store stores.Store) {
	switch r.Method {
	case http.MethodGet:
		urls, err := store.GetURLs()
		if err != nil {
			println(err.Error())
			http.Error(w, "Failed to fetch URLs", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(urls)
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			println(err.Error())
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var decodedBody newURLRequest
		err = json.Unmarshal(body, &decodedBody)
		if err != nil {
			println(err.Error())
			http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
			return
		}

		if decodedBody.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(decodedBody.Password), bcrypt.DefaultCost)
			if err != nil {
				println(err.Error())
				http.Error(w, "Failed to salt password", http.StatusInternalServerError)
				return
			}
			(&decodedBody).Password = string(hashedPassword)
		}

		if decodedBody.Slug == "" {
			slugLength := decodedBody.SlugLength
			if decodedBody.SlugLength == 0 {
				slugLength = 5
			}

			slug, err := generateSlug(slugLength)
			if err != nil {
				println(err.Error())
				http.Error(w, "Failed to generate slug", http.StatusInternalServerError)
				return
			}
			decodedBody.Slug = slug
		} else {
			url, _ := store.GetURL(decodedBody.Slug)
			if url != nil {
				http.Error(w, "Slug already exists", http.StatusConflict)
				return
			}
		}

		inserted, err := store.InsertURL(decodedBody.Slug, decodedBody.URL, decodedBody.Password, decodedBody.AllowedVisits)
		if err != nil {
			println(err.Error())
			http.Error(w, "Failed to save URL", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(inserted)
	}
}

func urlHandler(w http.ResponseWriter, r *http.Request, store stores.Store) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	if slug == "" {
		http.Error(w, "Invalid slug", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		url, err := store.GetURL(slug)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if url == nil {
			http.Error(w, "No URL found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(url)
	case http.MethodDelete:
		err := store.DeleteURL(slug)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			println(err.Error())
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var newURL updateURLRequest
		err = json.Unmarshal(body, &newURL)
		if err != nil {
			println(err.Error())
			http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
			return
		}

		if newURL.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newURL.Password), bcrypt.DefaultCost)
			if err != nil {
				println(err.Error())
				http.Error(w, "Failed to salt password", http.StatusInternalServerError)
				return
			}
			(&newURL).Password = string(hashedPassword)
		}

		err = store.UpdateURL(slug, newURL.URL, newURL.Password, newURL.AllowedVisits)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// SetUpUrlsHandlers - set up the /urls REST handlers
func SetUpUrlsHandlers(subrouter *mux.Router) error {
	store, err := stores.StoreFactory(config.Config.StoreType)
	if err != nil {
		return err
	}

	subrouter.HandleFunc("/{slug}", func(w http.ResponseWriter, r *http.Request) {
		urlHandler(w, r, store)
	}).Methods("GET", "PUT", "DELETE")

	subrouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		urlsHandler(w, r, store)
	}).Methods("GET", "POST")

	return nil
}
