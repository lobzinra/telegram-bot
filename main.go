package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	apiURL = "https://api.telegram.org/bot%s/%s"
)

var (
	apiToken   string
	httpClient *http.Client
	offset     int
)

func init() {
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
}

func main() {
	if len(os.Args) < 2 {
		panic(fmt.Errorf("ApiToken not found"))
	}

	apiToken = os.Args[1]

	ch := getUpdatesChan()

	for update := range ch {
		proccess(update)
	}
}

func getMe() (User, error) {
	const methodName = "getMe"

	user := User{}

	r, err := sendRequest(methodName, nil)
	if err != nil {
		return user, err
	}

	err = json.Unmarshal(r.Result, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func getUpdates(offset int) ([]Update, error) {
	const methodName = "getUpdates"

	updates := []Update{}

	data := url.Values{}

	if offset > 0 {
		data.Add("offset", strconv.Itoa(offset))
	}

	r, err := sendRequest(methodName, data)
	if err != nil {
		return updates, err
	}

	err = json.Unmarshal(r.Result, &updates)
	if err != nil {
		return updates, err
	}

	return updates, nil
}

func getUpdatesChan() chan Update {
	ch := make(chan Update)
	go func() {
		for {
			updates, err := getUpdates(offset)
			if err != nil {
				logError(err)
				time.Sleep(time.Second * 5)
				continue
			}

			for _, update := range updates {
				offset = update.ID + 1
				ch <- update
			}

			time.Sleep(time.Second * 1)
		}
	}()
	return ch
}

func sendMessage(msg MessageData) (bool, error) {
	const methodName = "sendMessage"

	data := url.Values{}
	data.Add("chat_id", strconv.Itoa(msg.ChatID))
	data.Add("text", msg.Text)
	data.Add("parse_mode", msg.ParseMode)
	data.Add("reply_to_message_id", strconv.Itoa(msg.ReplyTo))

	_, err := sendRequest(methodName, data)
	if err != nil {
		return false, err
	}

	return true, nil
}

func proccess(update Update) {
	msg := MessageData{
		ChatID:    update.Message.Chat.ID,
		Text:      fmt.Sprintf("Hi, %s! I'm bot", update.Message.From.FirstName),
		ParseMode: "HTML",
		ReplyTo:   update.Message.ID,
	}

	_, err := sendMessage(msg)
	if err != nil {
		logError(err)
	}
}

func sendRequest(methodName string, data url.Values) (ApiResponse, error) {
	r := ApiResponse{}

	url := fmt.Sprintf(apiURL, apiToken, methodName)

	resp, err := httpClient.PostForm(url, data)
	if err != nil {
		return r, err
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return r, err
	}

	return r, nil
}

func logError(err error) {
	fmt.Printf("ERROR: %v\n", err)
}

type ApiResponse struct {
	Ok          bool
	Description string
	Result      json.RawMessage
}

type Update struct {
	ID      int `json:"update_id"`
	Message Message
}

type Message struct {
	ID   int `json:"message_id"`
	From User
	Chat Chat
	Text string
}

type User struct {
	ID        int
	FirstName string `json:"first_name"`
}

type Chat struct {
	ID int
}

type MessageData struct {
	ChatID    int
	Text      string
	ParseMode string
	ReplyTo   int
}
