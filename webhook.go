package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type WebhookMessage struct {
	Content string `json:"content"`
	Username string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
}

func SendToWebhook(str string) error {
	m, _ := json.Marshal(WebhookMessage{
		Content: str,
		Username: "hexFS",
		AvatarUrl: "https://i.imgur.com/JDqEwet.png",
	})
	res, err := http.Post(os.Getenv(DiscordWebhookURL), "application/json", bytes.NewBuffer(m))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}
