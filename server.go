package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
)

func loadFile(path string) []byte {
	dat, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("File not found at:", path)
	}
	return dat
}

var atproto_redirect_string = "https://?"

func getRandBytes(n int) []byte {
	dat := make([]byte, n)
	_, err := rand.Read(dat)
	if err != nil {
		panic(1)
	}
	return dat
}

var discord_endpoints = oauth2.Endpoint{
	AuthURL:  "https://discord.com/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	conf := &oauth2.Config{
		ClientID:     cfg["discordClientID"].(string),
		ClientSecret: cfg["discordClientSecret"].(string),
		Scopes:       []string{"identify", "guilds"},
		Endpoint:     discord_endpoints,
	}
	dat := getRandBytes(16)
	verf := oauth2.GenerateVerifier()
	state := base64.URLEncoding.EncodeToString(dat)
	enc, _ := sc.Encode("oauth_meta", map[string]string{
		"state":    state,
		"verf":     verf,
		"redirect": r.URL.Query().Get("redirect"),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_meta",
		Value:    enc,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
	})
	url := conf.AuthCodeURL(state, oauth2.S256ChallengeOption(verf))
	tmpl := template.Must(template.ParseFiles("./static/login.html"))
	tmpl.Execute(w, map[string]string{
		"DiscordRedirect": url,
		"ATProtoRedirect": "",
	})
}
func badState(w http.ResponseWriter, r *http.Request, original_dest string) {
	if cfg["subdomain"] != nil {
		http.Redirect(w, r, fmt.Sprintf("https://%s.%s/login?redirect=%s", cfg["subdomain"], cfg["domainName"], original_dest), http.StatusFound)
	} else {
		http.Redirect(w, r, fmt.Sprintf("https://%s/login?redirect=%s", cfg["domainName"], original_dest), http.StatusFound)
	}
}
func authHandler(w http.ResponseWriter, r *http.Request) {
	host := r.Header.Get("X-Forwarded-Host")
	uri := r.Header.Get("X-Forwarded-Uri")
	original_dest := "https://" + host + uri
	cookie, err := r.Cookie("token")

	if err != http.ErrNoCookie {
		var auth map[string]interface{}
		err = sc.Decode("token", cookie.Value, &auth)
		if err != nil {
			badState(w, r, original_dest)
		}
		w.WriteHeader(http.StatusOK)
	}
	badState(w, r, original_dest)
}

var sc = securecookie.New(
	getRandBytes(32),
	getRandBytes(32),
)

func errChk(err error) {
	if err != nil {
		log.Println(err)
	}
}
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	from := r.PathValue("provider")
	if from == "discord" {
		cookie, err := r.Cookie("oauth_meta")
		errChk(err)
		var oauth_meta map[string]string
		sc.Decode("oauth_meta", cookie.Value, &oauth_meta)
		if oauth_meta["state"] != r.URL.Query().Get("state") {
			return
		}
		payload := url.Values{}
		payload.Set("grant_type", "authorization_code")
		payload.Set("code", r.URL.Query()["code"][0])
		payload.Set("code_verifier", oauth_meta["verf"])

		req, _ := http.NewRequest("POST", discord_endpoints.TokenURL, strings.NewReader(payload.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(cfg["discordClientID"].(string), cfg["discordClientSecret"].(string))
		resp, _ := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		var res map[string]interface{}

		json.NewDecoder(resp.Body).Decode(&res)
		req, err = http.NewRequest("GET", "https://discord.com/api/users/@me/guilds", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", res["access_token"]))
		guildsResp, err := http.DefaultClient.Do(req)
		var guilds []struct {
			ID string `json:"id"`
		}
		json.NewDecoder(guildsResp.Body).Decode(&guilds)
		enc, err := sc.Encode("token", res)
		for _, g := range guilds {
			if cfg["discordGuildID"].(string) == g.ID {
				http.SetCookie(w, &http.Cookie{
					Name:     "token",
					Value:    enc,
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
					Domain:   cfg["domainName"].(string),
					MaxAge:   300,
				})
				http.Redirect(w,r, oauth_meta["redirect"],http.StatusFound)
			}
		}

	}
}

var cfg map[string]interface{}

func main() {

	if err := json.Unmarshal(loadFile("./config.json"), &cfg); err != nil {
		panic(err)
	}
	fmt.Println(cfg)
	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/callback/{provider}", callbackHandler)
	fmt.Println("listening on port:", cfg["port"])
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg["port"].(string)), nil))
}
