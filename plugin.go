package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	config    *Config
	requester *Requester
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
	p.requester = NewRequester()
	log.Println("Gotify host: ", p.config.GotifyHost)
	// log.Println("Client token: ", p.config.ClientToken)
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
	time.Sleep(3 * time.Second)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err) // TODO: Make this retry instead of panic
	}

	go func(c *websocket.Conn) {
		var currentMessage PluginMessage
		defer c.Close()
		for {
			select {
			case <-interrupt:
				log.Println("received interrupt, closing connection")
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("read error:", err)
					return
				}
				if err := json.Unmarshal(message, &currentMessage); err != nil {
					log.Println("unable to unmarshal message:", err)
					continue
				}

				log.Printf("received: %s\n", message)
				for _, webhook := range p.config.Webhooks {
					log.Println("current app id:", currentMessage.AppId, "webhook app id:", webhook.AppId)
					if webhook.AppId == currentMessage.AppId {
						if err := p.SendToWebhook(webhook, currentMessage.Message); err != nil {
							log.Println("unable to send message to webhook:", err)
						}
					}
				}
			}
		}
	}(c)
}

func (p *Plugin) SendToWebhook(webhook *Webhook, message plugin.Message) error {
	requestBody := Message{Content: message.Message}

	response, err := p.requester.Post(context.Background(), webhook.Url, requestBody, nil)
	if err != nil {
		return err
	}

	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected response status code: %d", response.StatusCode)
	}

	return nil
}

func main() {
	panic("this should be built as go plugin")
}
