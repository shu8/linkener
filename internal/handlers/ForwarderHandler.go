package handlers

import (
	"net/http"
	"url-shortener/internal/stores"
	"text/template"

	"golang.org/x/crypto/bcrypt"
)

func redirect(w http.ResponseWriter, r *http.Request, store stores.Store, url *stores.ShortURL) {
	err := store.RecordVisit(url.Slug)
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to unshorten URL", http.StatusInternalServerError)
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	http.Redirect(w, r, url.URL, http.StatusMovedPermanently)
}

// ForwarderHandler - perform the short URL HTTP redirects on the / route
func ForwarderHandler(w http.ResponseWriter, r *http.Request, storeType string) {
	store, err := stores.StoreFactory(storeType)
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to unshorten URL", http.StatusBadRequest)
		return
	}

	slug := r.URL.Path[1:]
	url, err := store.GetURL(slug)

	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to unshorten URL", http.StatusInternalServerError)
		return
	}

	if url == nil {
		http.Error(w, "Unknown URL", http.StatusNotFound)
		return
	}

	if url.Stats.Count >= url.AllowedVisits {
		http.Error(w, "URL expired", http.StatusForbidden)
		return
	}

	if url.Password != "" {
		tmpl := template.Must(template.ParseFiles("urlPassword.html"))
		if r.Method != http.MethodPost {
			tmpl.Execute(w, struct{ Error bool }{false})
		} else {
			password := r.FormValue("password")
			if bcrypt.CompareHashAndPassword([]byte(url.Password), []byte(password)) != nil {
				tmpl.Execute(w, struct{ Error bool }{true})
			} else {
				redirect(w, r, store, url)
			}
		}
	} else {
		redirect(w, r, store, url)
	}
}
