package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"git.omegaice.dev/atitd/research-bot/chatlog"
)

/*
Egypt has unlocked the secrets of Stonework!
*/
func onTechComplete(m *chatlog.Message) {
	if !strings.Contains(m.Message, "Egypt has unlocked ") {
		return
	}

	re := regexp.MustCompile(`Egypt has unlocked .* secrets of (.*)!`)
	parts := re.FindAllStringSubmatch(m.Message, -1)
	if len(parts) == 0 {
		log.Println(len(parts))
		return
	}

	if len(parts[0]) != 2 {
		log.Println(len(parts[0]), parts[0])
		return
	}

	qTech, err := SemanticQuery(fmt.Sprintf("[[Category:Technologies]] [[~Technology/%s/*]] [[Is progressing::true]] |?Has level |sort=Has level |order=asc", parts[0][1]))
	if err != nil {
		log.Println(err)
		return
	}

	page := ""
	for key := range qTech.From("query.results.[0]").Get().(map[string]interface{}) {
		page = key
	}

	// Download and update the page
	current, err := DownloadPage(page)
	if err != nil {
		log.Println(err)
		return
	}

	newText := strings.Replace(strings.Replace(current, "  |complete=no", fmt.Sprintf("  |ended=%s\n  |complete=yes", time.Now().Format((time.RFC3339))), 1), "  |inprogress=yes", "  |inprogress=no", 1)

	resp, err := ModifyPage(page, newText, false)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(resp)

	qLabs, err := SemanticQuery("[[Category:Laboratories]]")
	if err != nil {
		log.Println(err)
		return
	}

	for key := range qLabs.From("query.results.[0]").Get().(map[string]interface{}) {
		if _, err := MarkLabComplete(key, parts[0][1]); err != nil {
			log.Println(err)
		}
	}
}
