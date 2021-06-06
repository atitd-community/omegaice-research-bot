package chatlog

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func New() *API {
	return &API{Handlers: make(map[string][]func(*Message))}
}

type API struct {
	Handlers map[string][]func(*Message)
}

func (a *API) AddHandler(channel string, function func(*Message)) {
	a.Handlers[channel] = append(a.Handlers[channel], function)
}

func (a *API) Tail() {
	for {
		for channel := range a.Handlers {
			time.Sleep(1 * time.Minute)

			messages, err := a.Download(channel, time.Now())
			if err != nil {
				log.Println(err)
				continue
			}

			for _, handler := range a.Handlers[channel] {
				for _, message := range messages {
					handler(&message)
				}
			}
		}
	}
}

func (a *API) Download(channel string, timestamp time.Time) ([]Message, error) {
	resp, err := http.Get(fmt.Sprintf("https://logs.atitd.wiki/api/%s/%d", channel, timestamp.Unix()))
	if err != nil {
		return []Message{}, err
	}

	var results []Message
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return []Message{}, err
	}

	return results, nil
}
