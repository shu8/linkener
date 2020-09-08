package main

import (
	"database/sql"
	"url-shortener/internal/db"
	"url-shortener/internal/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	storeType  = "json"
	privateAPI = false
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		next.ServeHTTP(w, r)
	})
}

func main() {
	var err error
	db.DBCon, err = sql.Open("sqlite3", "auth.db")
	if err != nil {
		log.Fatal("Failed to open auth database")
		db.DBCon.Close()
		return
	}

	router := mux.NewRouter()
	router.Use(corsMiddleware)

	api := router.PathPrefix("/api").Subrouter()

	urls := api.PathPrefix("/urls").Subrouter()
	urls.Use(handlers.AuthMiddleware)
	err = handlers.SetUpUrlsHandlers(urls, storeType)
	if err != nil {
		log.Fatal("Error starting /urls: " + err.Error())
	}

	auth := api.PathPrefix("/auth").Subrouter()
	err = handlers.SetUpAuthHandlers(auth)
	if err != nil {
		log.Fatal("Error starting /auth: " + err.Error())
	}

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.ForwarderHandler(w, r, storeType)
	})

	if privateAPI {
		log.Fatal(http.ListenAndServe("127.0.0.1:3000", router))
	} else {
		log.Fatal(http.ListenAndServe(":3000", router))
	}
	db.DBCon.Close()
}
