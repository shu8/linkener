package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"linkener/internal/config"
	"linkener/internal/db"
	"linkener/internal/handlers"
	"log"
	"net/http"
	"os/user"
	"path/filepath"

	"github.com/gorilla/mux"
)

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, DNT, Referer, User-Agent")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)
		next.ServeHTTP(w, r)
	})
}

func main() {
	user, err := user.Current()
	if err != nil {
		log.Fatal("Failed to find config file: " + err.Error())
		return
	}

	var configFileLocation string
	defaultConfigFileLocation := filepath.Join(user.HomeDir, ".linkener", "config.json")
	flag.StringVar(&configFileLocation, "config", defaultConfigFileLocation, "location for the config JSON file")
	flag.StringVar(&configFileLocation, "c", defaultConfigFileLocation, "location for the config JSON file (shorthand)")
	flag.Parse()

	configContents, err := ioutil.ReadFile(configFileLocation)
	if err != nil {
		log.Fatal("Failed to open config file: " + err.Error())
		return
	}

	err = json.Unmarshal(configContents, &config.Config)
	if err != nil {
		log.Fatal("Failed to open config file: " + err.Error())
		return
	}

	db.DBCon, err = sql.Open("sqlite3", config.Config.AuthDBLocation)
	if err != nil {
		db.DBCon.Close()
		log.Fatal("Failed to open auth database: " + err.Error())
		return
	}

	router := mux.NewRouter()

	api := router.PathPrefix("/api").Subrouter()
	api.Use(corsMiddleware)
	api.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)
	})

	urls := api.PathPrefix("/urls").Subrouter()
	// TODO uncomment
	// urls.Use(handlers.AuthMiddleware)
	err = handlers.SetUpUrlsHandlers(urls)
	if err != nil {
		log.Fatal("Error starting /urls: " + err.Error())
	}

	auth := api.PathPrefix("/auth").Subrouter()
	err = handlers.SetUpAuthHandlers(auth)
	if err != nil {
		log.Fatal("Error starting /auth: " + err.Error())
	}

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.ForwarderHandler(w, r)
	})

	fmt.Printf("Listening on port %d\n", config.Config.Port)
	if config.Config.PrivateAPI {
		log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", config.Config.Port), router))
	} else {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), router))
	}
	db.DBCon.Close()
}
