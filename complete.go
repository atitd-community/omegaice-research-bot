package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func onComplete(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != "850171092941537313" {
		return
	}

	if !strings.HasPrefix(m.Message.Content, "!complete") {
		return
	}

	re := regexp.MustCompile(`([-]?\d+)[,|\s]+([-]?\d+)\s+(.*)`)
	parts := re.FindAllStringSubmatch(m.Message.Content, -1)
	if len(parts) == 0 {
		log.Println(len(parts))
		return
	}

	if len(parts[0]) != 4 {
		log.Println(len(parts[0]), parts[0])
		return
	}

	// Check if the tech is in progress
	qProgress, err := SemanticQuery(fmt.Sprintf("[[~Technology/%s/*]] [[Is progressing::true]]", parts[0][3]))
	if err != nil {
		log.Println(err)
		return
	}

	rCount, err := qProgress.From("query.meta.count").GetR()
	if err != nil {
		log.Println(err)
		return
	}

	count, err := rCount.Uint()
	if err != nil {
		log.Println(err)
		return
	}

	if count == 0 {
		if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s is not currently in progress.", parts[0][3])); err != nil {
			log.Println(err)
		}
		return
	}

	// Find the correct lab
	qLabs, err := SemanticQuery(fmt.Sprintf("[[Category:Laboratories]] [[Has longitude::%s]] [[Has latitude::%s]]", parts[0][1], parts[0][2]))
	if err != nil {
		log.Println(err)
		return
	}

	rResult := qLabs.From("query.results.[0]").Get()
	if rResult == nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No labs found at %s, %s", parts[0][1], parts[0][2]))
		return
	}

	page := ""
	for key := range rResult.(map[string]interface{}) {
		page = key
	}

	// Download the page
	code, err := MarkLabComplete(page, parts[0][3])
	if err != nil {
		log.Println(err)
		return
	}

	if code == 1 {
		if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s is already complete at %s.", parts[0][3], page)); err != nil {
			log.Println(err)
		}
		return
	}

	if code == 2 {
		if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Updated %s to mark %s as completed.", page, parts[0][3])); err != nil {
			log.Println(err)
		}
	}
}

func MarkLabComplete(page string, technology string) (int, error) {
	// Download the page
	current, err := DownloadPage(page)
	if err != nil {
		return -1, err
	}

	if !strings.Contains(current, technology) {
		return 1, nil
	}

	var result []string
	skip := false
	for _, line := range strings.Split(current, "\n") {
		if !skip {
			if strings.Contains(line, technology) {
				skip = true
			} else {
				result = append(result, line)
			}
		} else {
			if strings.TrimSpace(line) == "}}" {
				skip = false
			}
		}
	}

	resp, err := ModifyPage(page, strings.Join(result, "\n"), false)
	if err != nil {
		return -1, err
	}
	log.Println(resp)

	return 2, nil
}
