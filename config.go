package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type config struct {
	DidAllowList        []string `json:"didAllowList"`
	DiscordGuildID      *string  `json:"discordGuildID"`
	DiscordClientID     *string  `json:"discordClientID"`
	DiscordClientSecret *string  `json:"discordClientSecret"`
	TailscaleSock       *string  `json:"tailscaleSock"`
	DomainName          *string  `json:"domainName"`
	Subdomain           *string  `json:"subdomain"`
	Port                *string  `json:"port"`
	Debug               *bool    `json:"debug"`
	AllowedDomains      []string `json:"allowedDomains"`
	HashKey             *string  `json:"hashKey"`
	BlockKey            *string  `json:"blockKey"`
}

// load precedence no overwrites
// ENV > config.json > generate
func loadConfig() config {

	var c config
	c.loadFromJson()
	c.loadFromEnv()
	c.validateConfig()
	return c

}
func (c *config) validateConfig() {

	if c.Port == nil {
		_, err := strconv.Atoi(*c.Port)
		if err != nil {
			log.Fatal("no port provided, unable to start server")
		}
	}
	if len(c.AllowedDomains) == 0 {
		log.Fatal("no allowed domains specified, server will not do anything")
	}
	if c.DomainName == nil {
		log.Fatal("domain name unspecified, unable to start server")
	}
	if c.DiscordClientID == nil && c.DiscordClientSecret == nil && c.DiscordGuildID == nil &&
		c.DidAllowList == nil &&
		c.TailscaleSock == nil {
		log.Fatal("no way to auth specified, server will not do anything")
	}
	if c.Debug == nil {
		c.Debug = new(bool)
		*c.Debug = false
	}
}
func (c *config) loadFromEnv() {
	didList := os.Getenv("PHRAG_DID_ALLOW_LIST")
	if didList != "" {
		c.DidAllowList = strings.Split(didList, ",")
	}
	discordGuild := os.Getenv("PHRAG_DISCORD_GUILD_ID")
	if discordGuild != "" {
		c.DiscordGuildID = &discordGuild
	}
	discordClientID := os.Getenv("PHRAG_DISCORD_CLIENT_ID")
	if discordClientID != "" {
		c.DiscordClientID = &discordClientID
	}
	discordClientSecret := os.Getenv("PHRAG_DISCORD_CLIENT_SECRET")
	if discordClientSecret != "" {
		c.DiscordClientSecret = &discordClientSecret
	}
	tailscale := os.Getenv("PHRAG_TAILSCALE_SOCK")
	if tailscale != "" {
		c.TailscaleSock = &tailscale
	}
	domain := os.Getenv("PHRAG_DOMAIN_NAME")
	if domain != "" {
		c.DomainName = &domain
	}
	subdomain := os.Getenv("PHRAG_SUBDOMAIN")
	if subdomain != "" {
		c.Subdomain = &subdomain
	}
	port := os.Getenv("PHRAG_PORT")
	if port != "" {
		c.Port = &port
	}
	debug := os.Getenv("PHRAG_DEBUG")
	if debug != "" {
		b, _ := strconv.ParseBool(debug)
		// if parseBool throws an error it returns false, which is ok for debug
		c.Debug = new(bool)
		*c.Debug = b
	}
	hash := os.Getenv("PHRAG_HASH_KEY")
	if hash != "" {
		c.HashKey = &hash
	}
	block := os.Getenv("PHRAG_BLOCK_KEY")
	if block != "" {
		c.BlockKey = &block
	}
	allowedDomains := os.Getenv("PHRAG_ALLOWED_DOMAINS")
	if allowedDomains != "" {
		c.AllowedDomains = strings.Split(allowedDomains, ",")
	}
}

func (c *config) loadFromJson() {
	configPath := "./config.json"
	dat, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(dat, &c)
	if err != nil {
		fmt.Println(err)
	}
}
