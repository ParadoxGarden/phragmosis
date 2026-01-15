package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

func loadFile(path string) []byte {
	dat, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("File not found at:", path)
	}
	return dat
}

var discord_redirect_string = "https://discord.com/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=https://%s/good&scope=guilds"
var atproto_redirect_string = "https://?"

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/login.html"))
	tmpl.Execute(w, map[string]string{
		"DiscordRedirect": fmt.Sprintf(discord_redirect_string, cfg["discord_client_id"], cfg["hostname"]),
		"ATProtoRedirect": "",
	})

}
func authHandler(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Authorization") == "" {
		host := r.Header.Get("X-Forwarded-Host")
		uri := r.Header.Get("X-Forwarded-Uri")
		original_dest := "https://" + host + uri
		http.Redirect(w, r, fmt.Sprintf("https://%s/login?redirect=%s", cfg["hostname"], original_dest), http.StatusFound)

	}
}

// callback url
func goodHandler(w http.ResponseWriter, r *http.Request) {
	from := r.Header.Get("Referer")
	if from == "https://discord.com/" { //highly opinionated
		//oauth !
		return
	}else{ // this is probably atproto but i'll get that exact url later
 		return
	}
}
func badHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(loadFile("./static/bad.html"))
}

var cfg map[string]interface{}

func main() {

	if err := json.Unmarshal(loadFile("./config.json"), &cfg); err != nil {
		panic(err)
	}
	fmt.Println(cfg)
	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/good", goodHandler)
	http.HandleFunc("/bad", badHandler)
	fmt.Println("listening on port:", cfg["port"])
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg["port"].(string)), nil))
}
