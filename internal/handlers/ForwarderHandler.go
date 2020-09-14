package handlers

import (
	"net/http"
	"text/template"
	"url-shortener/internal/config"
	"url-shortener/internal/stores"

	"golang.org/x/crypto/bcrypt"
)

var passwordTemplate = `<html><head><title>Password required!</title></head>
<body>
    {{if .Error}}
    <h2>Incorrect password!</h2>
    {{end}}

    <form method="POST">
		<label>Password:</label><br />
		<input type=hidden name=referer value={{.Referer}}>
        <input type=password name=password required><br />
        <input type=submit>
    </form>
</body></html>
`

type templateData struct {
	Error   bool
	Referer string
}

func redirect(w http.ResponseWriter, r *http.Request, store stores.Store, url *stores.ShortURL, referer string) {
	// TODO add more stats like location?
	err := store.RecordVisit(url.Slug, referer)
	if err != nil {
		println(err.Error())
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

	if url.AllowedVisits > 0 && len(url.Visits) >= url.AllowedVisits {
		http.Error(w, "URL expired", http.StatusForbidden)
		return
	}

	if url.Password != "" {
		tmpl := template.Must(template.New("passwordTemplate").Parse(passwordTemplate))
		if r.Method != http.MethodPost {
			tmpl.Execute(w, templateData{
				Error:   false,
				Referer: r.Referer(),
			})
		} else {
			password := r.FormValue("password")
			referer := r.FormValue("referer")
			if bcrypt.CompareHashAndPassword([]byte(url.Password), []byte(password)) != nil {
				tmpl.Execute(w, templateData{
					Error:   true,
					Referer: referer,
				})
			} else {
				redirect(w, r, store, url, referer)
			}
		}
	} else {
		redirect(w, r, store, url, r.Referer())
	}
}
