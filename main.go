package main

import (
	"log"
	"net/http"

	"forum/database"
	"forum/handler"
)

func main() {

	database.InitDB()
	defer database.DB.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		handler.IndexHandler(w, r, "index")
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.LoginGET(w, r)
		case http.MethodPost:
			handler.LoginPOST(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.RegisterGET(w, r)
		case http.MethodPost:
			handler.RegisterPOST(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.LogoutPOST(w, r)
	})

	log.Println("Listening on :8080 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
