package main

import (
	"net/http"

	"github.com/gotify/plugin-api"
)

type Config struct {
	GotifyHost  string     `json:"gotify_host"`
	ClientToken string     `json:"client_token"`
	Webhooks    []*Webhook `json:"webhooks"`
}

type Webhook struct {
	AppId int    `json:"app_id"`
	Name  string `json:"name"`
	Url   string `json:"url"`
}

type HttpResponse struct {
	Status     string
	StatusCode int
	Header     http.Header
	Body       []byte
}

type Message struct {
	Content string `json:"content"`
}

type PluginMessage struct {
	plugin.Message
	AppId int `json:"appid"`
}
