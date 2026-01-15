package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world!")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", helloWorldHandler).Methods("GET")
	http.HandleFunc("/", helloWorldHandler)

	fmt.Println("running server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
