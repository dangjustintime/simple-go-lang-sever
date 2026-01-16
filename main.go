package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type App struct {
	db *sql.DB
}

func main() {
	// connect to database
	db, err := sql.Open("sqlite", "users.db")
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return
	}

	app := &App{db: db}

	// ensure users table exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		name TEXT PRIMARY KEY,
		password TEXT
	);`)
	if err != nil {
		fmt.Println("Failed to ensure users table exists:", err)
		return
	}

	// set up router
	router := mux.NewRouter()
	router.HandleFunc("/users", app.getUsersHandler).Methods("GET")
	router.HandleFunc("/users", app.createUserHandler).Methods("POST")
	http.Handle("/", router)

	fmt.Println("running server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

func (app *App) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := app.db.Query("SELECT * FROM users;")
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Username, &u.Password); err != nil {
			http.Error(w, "Failed to scan user", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Row iteration error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (app *App) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(u.Username) == "" || strings.TrimSpace(u.Password) == "" {
		http.Error(w, "name and password are required", http.StatusBadRequest)
		return
	}

	_, err := app.db.Exec("INSERT INTO users(username, password) VALUES(?, ?);", u.Username, u.Password)
	if err != nil {
		// best-effort check for unique constraint
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "constraint") {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}
