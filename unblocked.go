package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func onUnblocked(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != "850171092941537313" {
		return
	}

	if !strings.HasPrefix(m.Message.Content, "!unblock") && !strings.HasPrefix(m.Message.Content, "!unblocked") {
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
	current, err := DownloadPage(page)
	if err != nil {
		log.Println(err)
		return
	}

	if !strings.Contains(current, parts[0][3]) {
		if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s is not currently in progress.", parts[0][3])); err != nil {
			log.Println(err)
		}
		return
	}

	var result []string
	skip := false
	alreadyUnblocked := true
	for _, line := range strings.Split(current, "\n") {
		if !skip {
			result = append(result, line)
			if strings.Contains(line, parts[0][3]) {
				skip = true
			}
		} else {
			if strings.TrimSpace(line) == "}}" {
				skip = false
				result = append(result, "}}")
			} else {
				if len(strings.TrimSpace(line)) != 0 {
					alreadyUnblocked = false
				}
			}
		}
	}

	if !alreadyUnblocked {
		resp, err := ModifyPage(page, strings.Join(result, "\n"), false)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(resp)

		if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Updated %s to mark %s as unblocked.", page, parts[0][3])); err != nil {
			log.Println(err)
		}
	} else {
		if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s is already marked as unblocked at %s.", parts[0][3], page)); err != nil {
			log.Println(err)
		}
	}
}
