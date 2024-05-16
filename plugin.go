package main

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/gotify/plugin-api"
)

// GetGotifyPluginInfo returns gotify plugin info
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		ModulePath:  "github.com/ebaldebo/gotify-webhook",
		Name:        "gotify-webhook",
		Version:     "0.1.0",
		Author:      "ebaldebo",
		License:     "MIT",
		Description: "Forward messages to webhook(Discord, Slack, etc.)",
	}
}

// Plugin is plugin instance
type Plugin struct {
	config *Config
}

func (p *Plugin) DefaultConfig() interface{} {
	return &Config{}
}

func (p *Plugin) ValidateAndSetConfig(c interface{}) error {
	config := c.(*Config)
	p.config = config
	return nil
}

// Enable implements plugin.Plugin
func (p *Plugin) Enable() error {
	if p.config.GotifyHost == "" {
		return errors.New("gotify host is required")
	}
	log.Println("Gotify host: ", p.config.GotifyHost)
	log.Println("Client token: ", p.config.ClientToken)
	for _, webhook := range p.config.Webhooks {
		log.Println("Webhook: ", webhook.AppId, webhook.Name, webhook.Url)
	}

	go p.HandleMessages()

	return nil
}

// Disable implements plugin.Plugin
func (p *Plugin) Disable() error {
	return nil
}

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &Plugin{}
}

func (p *Plugin) HandleMessages() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(interrupt, syscall.SIGTERM)

	url := p.config.GotifyHost + "/stream?token=" + p.config.ClientToken
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}

	go func(c *websocket.Conn) {
		defer c.Close()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}
			log.Printf("received: %s\n", message)
		}
	}(c)
}

func main() {
	panic("this should be built as go plugin")
}
