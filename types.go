package main

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
