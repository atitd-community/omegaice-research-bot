package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.omegaice.dev/atitd/research-bot/chatlog"
)

/*
Scientists at the University of The Human Body have been instructed to commission investigation into Dowsing by House Hyksos; they will report back on their progress in due course and any further requirements individual laboratories may have.
*/
func onTechStart(m *chatlog.Message) {
	if !strings.Contains(m.Message, "have been instructed to commission investigation into") {
		return
	}

	re := regexp.MustCompile(`Scientists at the University of (.*) have.*into (.*) by House (.+);`)
	parts := re.FindAllStringSubmatch(m.Message, -1)
	if len(parts) == 0 {
		log.Println(len(parts))
		return
	}

	if len(parts[0]) != 4 {
		log.Println(len(parts[0]), parts[0])
		return
	}

	qTech, err := SemanticQuery(fmt.Sprintf("[[Category:Technologies]] [[~Technology/%s/*]] [[Has completed::false]] |?Has level |sort=Has level |order=asc", parts[0][2]))
	if err != nil {
		log.Println(err)
		return
	}

	page := ""
	currentLevel := 999
	for key := range qTech.From("query.results.[0]").Get().(map[string]interface{}) {
		kParts := strings.Split(key, "/")

		level, err := strconv.Atoi(kParts[2])
		if err != nil {
			log.Println()
			return
		}

		if level < currentLevel {
			currentLevel = level
			page = key
		}
	}

	points, err := CalculatePoints()
	if err != nil {
		log.Println(err)
		return
	}

	// Download the page
	current, err := DownloadPage(page)
	if err != nil {
		log.Println(err)
		return
	}

	newText := strings.Replace(current, "  |inprogress=no", fmt.Sprintf("  |started=%s\n  |inprogress=yes\n  |champion=%s\n  |points=%d", time.Now().Format((time.RFC3339)), strings.ToLower(parts[0][3]), points), 1)

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
		log.Println(key)
		if _, err := AddTechToLab(key, parts[0][2]); err != nil {
			log.Println(err)
		}
	}
}

func CalculatePoints() (int, error) {
	retVal := 5

	// Calculate complete increase
	qTechs, err := SemanticQuery("[[Category:Technologies]] [[Is generated::true]] [[Has completed::true]]")
	if err != nil {
		return -1, err
	}

	rCount, err := qTechs.From("query.meta.count").GetR()
	if err != nil {
		return -1, err
	}

	count, err := rCount.Int()
	if err != nil {
		return -1, err
	}

	retVal += count

	// Calculate scroll discount
	qScrolls, err := SemanticQuery("[[Scroll of Wisdom]]|?Turned in|mainlabel=-")
	if err != nil {
		return -1, err
	}

	rCount, err = qScrolls.From("query.results.Scroll of Wisdom.printouts.Turned in.[0]").GetR()
	if err != nil {
		return -1, err
	}

	count, err = rCount.Int()
	if err != nil {
		return -1, err
	}
	retVal -= count

	return retVal, nil
}

func AddTechToLab(page string, technology string) (int, error) {
	// Download the page
	current, err := DownloadPage(page)
	if err != nil {
		return -1, err
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

	result = append(result, fmt.Sprintf("{{LabTechnology|technology=%s|time=|", technology))
	result = append(result, "")
	result = append(result, "}}")

	resp, err := ModifyPage(page, strings.Join(result, "\n"), false)
	if err != nil {
		return -1, err
	}
	log.Println(resp)

	return 2, nil
}
