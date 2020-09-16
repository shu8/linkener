package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"github.com/shu8/linkener/internal/config"
	"github.com/shu8/linkener/internal/db"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type revokeRequest struct {
	AccessToken string `json:"access_token"`
}

type passwordChangeRequest struct {
	Password string `json:"password"`
}

func generateAccessToken() (string, error) {
	bytes := make([]byte, 50)

	_, err := rand.Read(bytes[:])
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), err
}

func editUserHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		println(err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var decodedBody passwordChangeRequest
	err = json.Unmarshal(body, &decodedBody)
	if err != nil {
		println(err.Error())
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	requestUsername := vars["username"]
	loggedInUsername := r.Context().Value(UsernameContextKey)
	if loggedInUsername == nil || loggedInUsername.(string) != requestUsername {
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(decodedBody.Password), bcrypt.DefaultCost)
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	stmt, err := db.DBCon.Prepare("update users set password=? where username=?")
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(string(hashedPassword), requestUsername)
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	http.ResponseWriter.Write(w, []byte("Success!"))
}

func newUserHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		println(err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var decodedBody authRequest
	err = json.Unmarshal(body, &decodedBody)
	if err != nil {
		println(err.Error())
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	stmt, err := db.DBCon.Prepare("select username from users where username=?")
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(decodedBody.Username)
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to add new user", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(decodedBody.Password), bcrypt.DefaultCost)
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to add new user", http.StatusInternalServerError)
		return
	}

	stmt, err = db.DBCon.Prepare("insert into users(username, password) values(?, ?)")
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to add new user", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(decodedBody.Username, string(hashedPassword))
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to add new user", http.StatusInternalServerError)
		return
	}

	http.ResponseWriter.Write(w, []byte("Success!"))
}

func generateTokenHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		println(err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var decodedBody authRequest
	err = json.Unmarshal(body, &decodedBody)
	if err != nil {
		println(err.Error())
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	stmt, err := db.DBCon.Prepare("select username, password from users where username=?")
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	var dbUsername, dbPassword string
	if stmt.QueryRow(decodedBody.Username).Scan(&dbUsername, &dbPassword) != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}

	if dbUsername == "" || dbPassword == "" {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(decodedBody.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, err := generateAccessToken()
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	stmt, err = db.DBCon.Prepare("insert or replace into access_tokens(username, access_token) values(?, ?)")
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(dbUsername, accessToken)
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	http.ResponseWriter.Write(w, []byte(accessToken))
}

func revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		println(err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var decodedBody revokeRequest
	err = json.Unmarshal(body, &decodedBody)
	if err != nil {
		println(err.Error())
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	stmt, err := db.DBCon.Prepare("delete from access_tokens where access_token=?")
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(decodedBody.AccessToken)
	if err != nil {
		println(err.Error())
		http.Error(w, "Failed to revoke access token", http.StatusInternalServerError)
		return
	}

	http.ResponseWriter.Write(w, []byte("Succesfully revoked access token"))
}

// AuthMiddleware - ensure valid access token is passed for API routes that require authentication
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized access", http.StatusUnauthorized)
			return
		}

		stmt, err := db.DBCon.Prepare("select username from access_tokens where access_token=? and expiry>CURRENT_TIMESTAMP limit 1")
		if err != nil {
			println(err.Error())
			http.Error(w, "Failed to authorize request", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		rows, err := stmt.Query(token)
		if err != nil {
			println(err.Error())
			http.Error(w, "Failed to authorize request", http.StatusInternalServerError)
			return
		}

		if rows.Next() {
			var username string
			rows.Scan(&username)
			rows.Close()
			ctxt := r.WithContext(context.WithValue(r.Context(), UsernameContextKey, username))
			next.ServeHTTP(w, ctxt)
		} else {
			rows.Close()
			http.Error(w, "Unauthorized access", http.StatusUnauthorized)
			return
		}
	})
}

// SetUpAuthHandlers - set up the /api/auth REST handlers
func SetUpAuthHandlers(subrouter *mux.Router) error {
	subrouter.Handle("/users/{username}", AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		editUserHandler(w, r)
	}))).Methods("PUT")

	if config.Config.RegistrationEnabled {
		subrouter.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
			newUserHandler(w, r)
		}).Methods("POST")
	}

	subrouter.HandleFunc("/new_token", func(w http.ResponseWriter, r *http.Request) {
		generateTokenHandler(w, r)
	}).Methods("POST")

	subrouter.HandleFunc("/revoke_token", func(w http.ResponseWriter, r *http.Request) {
		revokeTokenHandler(w, r)
	}).Methods("POST")

	return nil
}
