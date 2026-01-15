package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

type User struct {
	Name     string `json:"name"`
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

	// set up router
	router := mux.NewRouter()
	router.HandleFunc("/", helloWorldHandler).Methods("GET")
	router.HandleFunc("/users", app.getUsersHandler).Methods("GET")
	http.Handle("/", router)

	fmt.Println("running server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world!")
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
		if err := rows.Scan(&u.Name, &u.Password); err != nil {
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
