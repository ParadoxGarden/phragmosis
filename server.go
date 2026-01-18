package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
)

var atproto_redirect_string = "https://?"

func (s *server) loginHandler(w http.ResponseWriter, r *http.Request) {
	verf := oauth2.GenerateVerifier()
	state := base64.URLEncoding.EncodeToString(getRandBytes(16))
	redir := r.URL.Query().Get("redirect")
	redirect, err := url.Parse("https://" + redir)
	if err != nil {
		s.redirectLogin(w, r, "malformed redirect")
		return
	}
	goodDom := false
	for _, dom := range s.config.AllowedDomains {
		if redirect.Host == dom || strings.HasSuffix(redirect.Hostname(), "."+dom) {
			goodDom = true
		}
	}
	if !goodDom && !strings.HasPrefix(redir, "/") || strings.HasPrefix(redir, "//") {
		s.redirectLogin(w, r, "malformed redirect")
		return
	}
	enc, err := s.sc.Encode("oauthMeta", map[string]string{
		"state":    state,
		"verf":     verf,
		"redirect": redirect.String(),
	})
	if err != nil {
		s.redirectLogin(w, r, "error encoding oauth metadata")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthMeta",
		Value:    enc,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
	})
	discordURL := s.discordOauthConfig.AuthCodeURL(state, oauth2.S256ChallengeOption(verf))
	s.loginPageTemplate.Execute(w, map[string]string{
		"DiscordRedirect": discordURL,
		"ATProtoRedirect": "",
		"rzn":"rzn"
	})
}

func (s *server) redirectLogin(w http.ResponseWriter, r *http.Request, rzn string) {
	host := r.Header.Get("X-Forwarded-Host")
	uri := r.Header.Get("X-Forwarded-Uri")
	originalDest := host + uri
	if *s.config.Debug {
	http.Redirect(w, r, fmt.Sprintf("https://%s?redirect=%s&rzn=%s", s.selfDomain, originalDest,rzn), http.StatusFound)
	return
	}
	http.Redirect(w, r, fmt.Sprintf("https://%s?redirect=%s", s.selfDomain, originalDest), http.StatusFound)
}

func (s *server) authHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")

	if err == http.ErrNoCookie {
		s.redirectLogin(w, r, "")
		return
	}
	auth := &oauth2.Token{}
	err = s.sc.Decode("token", cookie.Value, &auth)
	if err != nil {
		s.redirectLogin(w, r, "token is old, please log in again")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) callbackHandler(w http.ResponseWriter, r *http.Request) {
	from := r.PathValue("provider")
	switch from {
	case "discord":
		cookie, err := r.Cookie("oauthMeta")
		if err != nil {
			s.redirectLogin(w, r, "no cookie for callback")
			return
		}
		var oauthMeta map[string]string
		code := r.FormValue("code")
		state := r.FormValue("state")
		err = s.sc.Decode("oauthMeta", cookie.Value, &oauthMeta)
		if err != nil {
			s.redirectLogin(w, r, "cookie failed to decode")
			return
		}
		if oauthMeta["state"] != state {
			s.redirectLogin(w, r, "state validation failed for CRSF protection")
			return
		}
		token, err := s.discordOauthConfig.Exchange(r.Context(), code, oauth2.VerifierOption(oauthMeta["verf"]))
		req, _ := http.NewRequest("GET", "https://discord.com/api/users/@me/guilds", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		cl := &http.Client{Timeout: time.Second * 10}
		guildsResp, err := cl.Do(req)
		if err != nil {
			s.redirectLogin(w, r, "discord api responded with: " + err.Error())
			return
		}
		var guilds []struct {
			ID string `json:"id"`
		}
		err = json.NewDecoder(guildsResp.Body).Decode(&guilds)
		if err != nil {
			s.redirectLogin(w, r, "discord api responded with: " + err.Error())
			return
		}
		enc, err := s.sc.Encode("token", token)
		if err != nil {
			s.redirectLogin(w, r, "error encoding token")
			return
		}
		for _, g := range guilds {
			if *s.config.DiscordGuildID == g.ID {
				http.SetCookie(w, &http.Cookie{
					Name:     "token",
					Value:    enc,
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
					Domain:   *s.config.DomainName,
					MaxAge:   604800,
				})
				http.Redirect(w, r, oauthMeta["redirect"], http.StatusFound)
				return
			}
		}
		s.redirectLogin(w, r, "user not in circle of trust")
		return

	default:
		s.redirectLogin(w, r, "")
		return
	}

}

type server struct {
	config             config
	sc                 securecookie.SecureCookie
	selfDomain         string
	loginPageTemplate  template.Template
	discordOauthConfig oauth2.Config
	discordEndpoints   oauth2.Endpoint
	discord            bool
	atproto            bool
}

func initServ(c config) *server {

	s := &server{}
	if c.DiscordGuildID != nil && c.DiscordClientID != nil && c.DiscordClientSecret != nil {
		s.discord = true
	} else {
		s.discord = false
	}
	if c.DidAllowList != nil {
		s.atproto = true
	} else {
		s.atproto = false
	}
	if s.discord {
		s.discordEndpoints = oauth2.Endpoint{
			AuthURL:  "https://discord.com/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		}
		s.discordOauthConfig = oauth2.Config{
			ClientID:     *c.DiscordClientID,
			ClientSecret: *c.DiscordClientSecret,
			Scopes:       []string{"identify", "guilds"},
			Endpoint:     s.discordEndpoints,
		}
	}
	s.loginPageTemplate = *template.Must(template.ParseFiles("./static/login.html"))
	if c.Subdomain != nil {
		s.selfDomain = fmt.Sprintf("%s.%s/", *c.Subdomain, *c.DomainName)
	} else {
		s.selfDomain = fmt.Sprintf("%s/", *c.DomainName)
	}
	s.sc = initSecureCookie(c.BlockKey, c.HashKey)
	s.config = c
	return s
}
func main() {
	c := loadConfig()
	s := initServ(c)
	http.HandleFunc("/", s.loginHandler)
	http.HandleFunc("/auth", s.authHandler)
	http.HandleFunc("/callback/{provider}", s.callbackHandler)
	fmt.Println("listening on port:", *s.config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *s.config.Port), nil))
}
