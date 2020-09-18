package handlers

import (
	"net/http"
	"text/template"

	"github.com/shu8/linkener/internal/config"
	"github.com/shu8/linkener/internal/static"
	"github.com/shu8/linkener/internal/stores"

	"golang.org/x/crypto/bcrypt"
)

type templateData struct {
	Password bool
	Unknown  bool
	Error    bool
	Expired  bool

	PasswordIncorrect bool

	Referer string
}

var tmpl = template.Must(template.New("passwordTemplate").Parse(static.PasswordTemplate))

func redirect(w http.ResponseWriter, r *http.Request, store stores.Store, url *stores.ShortURL, referer string) {
	// TODO add more stats like location?
	err := store.RecordVisit(url.Slug, referer)
	if err != nil {
		println(err.Error())
		tmpl.Execute(w, templateData{
			Error: true,
		})
		http.Error(w, "Failed to unshorten URL", http.StatusInternalServerError)
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	http.Redirect(w, r, url.URL, http.StatusMovedPermanently)
}

// ForwarderHandler - perform the short URL HTTP redirects on the / route
func ForwarderHandler(w http.ResponseWriter, r *http.Request) {
	store, err := stores.StoreFactory(config.Config.StoreType)
	if err != nil {
		println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, templateData{
			Error: true,
		})
		return
	}

	slug := r.URL.Path[1:]
	url, err := store.GetURL(slug)

	if err != nil {
		println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		tmpl.Execute(w, templateData{
			Error: true,
		})
		return
	}

	if url == nil {
		w.WriteHeader(http.StatusNotFound)
		tmpl.Execute(w, templateData{
			Unknown: true,
		})
		return
	}

	if url.AllowedVisits > 0 && len(url.Visits) >= url.AllowedVisits {
		w.WriteHeader(http.StatusForbidden)
		tmpl.Execute(w, templateData{
			Expired: true,
		})
		return
	}

	if url.Password != "" {
		if r.Method != http.MethodPost {
			tmpl.Execute(w, templateData{
				Password:          true,
				PasswordIncorrect: false,
				Referer:           r.Referer(),
			})
		} else {
			password := r.FormValue("password")
			referer := r.FormValue("referer")
			if bcrypt.CompareHashAndPassword([]byte(url.Password), []byte(password)) != nil {
				tmpl.Execute(w, templateData{
					Password:          true,
					PasswordIncorrect: true,
					Referer:           referer,
				})
			} else {
				redirect(w, r, store, url, referer)
			}
		}
	} else {
		redirect(w, r, store, url, r.Referer())
	}
}
