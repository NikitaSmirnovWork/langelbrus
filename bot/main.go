package main

// bot/main.go — запускается отдельно от API сервера
// go run bot/main.go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const apiBase = "https://api.telegram.org/bot"

var botToken = os.Getenv("BOT_TOKEN")
var webappURL = os.Getenv("WEBAPP_URL") // https://your-railway-app.up.railway.app

type Update struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

type GetUpdatesResp struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

func apiCall(method string, body any) ([]byte, error) {
	b, _ := json.Marshal(body)
	resp, err := http.Post(apiBase+botToken+"/"+method, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 512)
	for {
		n, e := resp.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if e != nil {
			break
		}
	}
	return buf, nil
}

func sendMessage(chatID int64, text string, replyMarkup any) {
	payload := map[string]any{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": replyMarkup,
	}
	apiCall("sendMessage", payload)
}

func handleUpdate(u Update) {
	if u.Message == nil {
		return
	}
	if u.Message.Text != "/start" {
		return
	}

	chatID := u.Message.Chat.ID
	keyboard := map[string]any{
		"inline_keyboard": [][]map[string]any{
			{
				{
					"text": "📚 Open Word Learner",
					"web_app": map[string]string{
						"url": webappURL,
					},
				},
			},
		},
	}
	sendMessage(chatID,
		"👋 Welcome to *Word Learner*!\n\nLearn English words organized by modules and topics.\nPress the button below to open the app.",
		keyboard,
	)
}

func main() {
	if botToken == "" {
		log.Fatal("BOT_TOKEN environment variable is required")
	}
	if webappURL == "" {
		log.Fatal("WEBAPP_URL environment variable is required")
	}

	log.Println("Bot started, polling for updates...")

	offset := 0
	for {
		raw, err := apiCall("getUpdates", map[string]any{
			"offset":  offset,
			"timeout": 30,
		})
		if err != nil {
			log.Println("getUpdates error:", err)
			time.Sleep(3 * time.Second)
			continue
		}

		var resp GetUpdatesResp
		if err := json.Unmarshal(raw, &resp); err != nil || !resp.OK {
			log.Println("Parse error:", string(raw))
			time.Sleep(3 * time.Second)
			continue
		}

		for _, u := range resp.Result {
			handleUpdate(u)
			offset = u.UpdateID + 1
		}
		fmt.Print(".") // heartbeat
	}
}
