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

	"github.com/ebaldebo/gotify-webhook/internal/connection"
	"github.com/ebaldebo/gotify-webhook/internal/requester"
	"github.com/gorilla/websocket"
	"github.com/gotify/plugin-api"
)

var req requester.Requester = requester.NewHttpRequester(&http.Client{Timeout: 5 * time.Second})
var con connection.Connection = connection.NewWebsocketConnection()

// GetGotifyPluginInfo returns gotify plugin info
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		ModulePath:  "github.com/ebaldebo/gotify-webhook",
		Name:        "gotify-webhook",
		Author:      "ebaldebo",
		License:     "MIT",
		Description: "Forward messages to webhook(Discord, Slack, etc.)",
	}
}

// Plugin is plugin instance
type Plugin struct {
	config     *Config
	requester  requester.Requester
	connection connection.Connection
	disable    chan struct{}
}

// Enable implements plugin.Plugin
func (p *Plugin) Enable() error {
	p.disable = make(chan struct{})
	go p.HandleMessages()

	return nil
}

// Disable implements plugin.Plugin
func (p *Plugin) Disable() error {
	close(p.disable)
	return nil
}

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &Plugin{
		requester:  req,
		connection: con,
	}
}

func (p *Plugin) DefaultConfig() interface{} {
	return &Config{
		GotifyHost: "wss://localhost",
	}
}

func (p *Plugin) ValidateAndSetConfig(c interface{}) error {
	config := c.(*Config)
	if config.ClientToken == "" {
		return errors.New("clienttoken is required")
	}

	for _, webhook := range config.Webhooks {
		if webhook.AppId == 0 {
			return errors.New("appid is required")
		}
	}

	p.config = config
	return nil
}

func (p *Plugin) HandleMessages() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	url := p.config.GotifyHost + "/stream?token=" + p.config.ClientToken

	c := p.connection.CreateWebsocketConnection(url)

	messageChannel := readMessages(c)
	p.processMessages(interrupt, messageChannel)
}

func readMessages(c *websocket.Conn) chan []byte {
	messageChannel := make(chan []byte)
	go func(c *websocket.Conn) {
		defer c.Close()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				close(messageChannel)
				return
			}
			messageChannel <- message
		}
	}(c)
	return messageChannel
}

func (p *Plugin) processMessages(interrupt chan os.Signal, messageChannel chan []byte) {
	var currentMessage PluginMessage
	for {
		select {
		case <-p.disable:
			return
		case <-interrupt:
			return
		case message, ok := <-messageChannel:
			if !ok {
				return
			}
			if err := json.Unmarshal(message, &currentMessage); err != nil {
				log.Println("unable to unmarshal message:", err)
				continue
			}

			log.Printf("received: %s\n", message)
			p.processWebhooks(currentMessage)
		}
	}
}

func (p *Plugin) processWebhooks(currentMessage PluginMessage) {
	for _, webhook := range p.config.Webhooks {
		if webhook.AppId == currentMessage.AppId {
			if err := p.sendToWebhook(webhook, currentMessage.Message); err != nil {
				log.Println("unable to send message to webhook:", err)
			}
		}
	}
}

func (p *Plugin) sendToWebhook(webhook *Webhook, message plugin.Message) error {
	requestBody := WebhookMessage{
		Title:   message.Title,
		Content: message.Message,
	}

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
