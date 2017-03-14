package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	ApiToken = "yourApiTokenHere"
	ApiUrl   = "https://api.telegram.org/bot%s/%s"
)

var (
	httpClient http.Client
)

func main() {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
		}
	}()

	httpClient = http.Client{}

	getMe()

	offset := 0

	ch := getUpdatesChan(offset)

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

func getUpdatesChan(offset int) chan Update {
	ch := make(chan Update)

	go func() {
		for {
			updates, err := getUpdates(offset)
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
				time.Sleep(time.Second * 5)
				continue
			}

			for _, update := range updates {
				offset = update.Id + 1
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
	data.Add("chat_id", strconv.Itoa(msg.ChatId))
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
		ChatId:    update.Message.Chat.Id,
		Text:      fmt.Sprintf("Hi, %s! I'm bot", update.Message.From.FirstName),
		ParseMode: "HTML",
		ReplyTo:   update.Message.Id,
	}

	_, err := sendMessage(msg)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}
}

func sendRequest(methodName string, data url.Values) (ApiResponse, error) {
	r := ApiResponse{}

	url := fmt.Sprintf(ApiUrl, ApiToken, methodName)

	resp, err := httpClient.PostForm(url, data)
	if err != nil {
		return r, err
	}

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

type ApiResponse struct {
	Ok          bool
	Description string
	Result      json.RawMessage
}

type Update struct {
	Id      int `json:"update_id"`
	Message Message
}

type Message struct {
	Id   int `json:"message_id"`
	From User
	Chat Chat
	Text string
}

type User struct {
	Id        int
	FirstName string `json:"first_name"`
}

type Chat struct {
	Id int
}

type MessageData struct {
	ChatId    int
	Text      string
	ParseMode string
	ReplyTo   int
}
