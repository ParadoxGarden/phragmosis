package main

import (
	// "fmt"
	"log"
	"net/http"
	"os"
)

func loadPageFromDisk(path string) []byte {
	dat, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("File not found at:", path)
	}
	return dat
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(loadPageFromDisk("./static/login.html"))
}

func authHandler(w http.ResponseWriter, r *http.Request) {

}
func badHandler(w http.ResponseWriter, r *http.Request) {

}

// need?
func goodHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/bad", badHandler)
	http.HandleFunc("/good", goodHandler) // again, need?
	log.Fatal(http.ListenAndServe(":8080", nil))
}
