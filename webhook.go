package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type WebhookMessage struct {
	Content string `json:"content"`
}

func SendToWebhook(str string) error {
	m, _ := json.Marshal(WebhookMessage{Content: str})
	res, err := http.Post(os.Getenv(DiscordWebhookURL), "application/json", bytes.NewBuffer(m))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}
