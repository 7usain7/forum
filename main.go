package main

import (
	"forum/database"
	"forum/handler"
	"log"
	"net/http"
)

func main() {

	database.InitDB()
	defer database.DB.Close()

	http.HandleFunc("/", handler.IndexHandler)
	http.HandleFunc("/login", handler.LoginHandler)
	http.HandleFunc("/register", handler.RegisterHandler)
	http.HandleFunc("/logout", handler.LogoutHandler)

	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Listening on :8080 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
