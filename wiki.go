package main

import (
	"log"

	"github.com/antonholmquist/jason"
	"github.com/thedevsaddam/gojsonq/v2"
)

func SemanticQuery(query string) (*gojsonq.JSONQ, error) {
	cParameters := map[string]string{
		"action": "ask",
		"query":  query,
		"format": "json",
	}

	cResp, err := Wiki.Get(cParameters)
	if err != nil {
		return nil, err
	}

	return gojsonq.New().FromString(cResp.String()), nil
}

func DownloadPage(page string) (string, error) {
	cParameters := map[string]string{
		"action": "parse",
		"prop":   "wikitext",
		"page":   page,
	}

	cResp, err := Wiki.Get(cParameters)
	if err != nil {
		return "", err
	}

	rCurrent, err := gojsonq.New().FromString(cResp.String()).From("parse.wikitext").GetR()
	if err != nil {
		return "", err
	}

	return rCurrent.String()
}

func ModifyPage(page string, text string, createonly bool) (string, error) {
	relog := false

	var err error
	for retry := 0; retry < 3; retry++ {
		if relog {
			if err = Wiki.Login(username, password); err != nil {
				return "", err
			}
			relog = false
		}

		var token string
		token, err = Wiki.GetToken("csrf")
		if err != nil {
			relog = true
			log.Println(err)
			continue
		}

		cParameters := map[string]string{
			"action": "edit",
			"bot":    "true",
			//"createonly": fmt.Sprintf("%t", createonly),
			"title": page,
			"text":  text,
			"token": token,
		}

		var cResp *jason.Object
		cResp, err = Wiki.Post(cParameters)
		if err != nil {
			relog = true
			log.Println(err)
			continue
		}

		return cResp.String(), nil
	}
	return "", err
}
