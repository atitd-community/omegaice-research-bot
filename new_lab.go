package main

import (
	"fmt"
	"log"
	"strings"

	"git.omegaice.dev/atitd/research-bot/chatlog"
	"github.com/thedevsaddam/gojsonq/v2"
)

//A new Laboratory has been completed at 995, 6813. Scientists require additional Laboratories to continue their research activities.
func onLabConstructed(m *chatlog.Message) {
	if !strings.Contains(m.Message, "A new Laboratory has been completed at ") {
		return
	}

	var lon int
	var lat int
	if _, err := fmt.Sscanf(m.Message, "A new Laboratory has been completed at %d, %d. Scientists require additional Laboratories to continue their research activities.", &lon, &lat); err != nil {
		return
	}

	cParameters := map[string]string{
		"action": "ask",
		"query":  "[[Category:Laboratories]]",
		"format": "json",
	}

	cResp, err := Wiki.Get(cParameters)
	if err != nil {
		log.Println(err)
		return
	}

	rCount, err := gojsonq.New().FromString(cResp.String()).From("query.meta.count").GetR()
	if err != nil {
		log.Println(err)
		return
	}

	count, err := rCount.Uint()
	if err != nil {
		log.Println(err)
		return
	}

	page := fmt.Sprintf("Laboratory/%d", count+1)
	wikiText := fmt.Sprintf(`{{Laboratory
  |name=
  |location={{Location | %d | %d}}
  |house=
}}`, lon, lat)

	token, err := Wiki.GetToken("csrf")
	if err != nil {
		log.Fatalln(err)
	}

	cParameters = map[string]string{
		"action":     "edit",
		"bot":        "true",
		"createonly": "true",
		"title":      page,
		"text":       wikiText,
		"token":      token,
	}

	cResp, err = Wiki.Post(cParameters)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(cResp)
}
