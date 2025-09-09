package parsers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type LoginInfo struct {
	Name     string
	Username string
}

func (s *service) ParseUserLogin(html string) (*LoginInfo, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	nameElement := doc.Find("#topnav span")
	if nameElement.Length() == 0 {
		if doc.Find("html body div:nth-child(2) div:nth-child(4) div span").Length() > 0 {
			return nil, fmt.Errorf("login rejected due to bad credentials")
		}
		return nil, fmt.Errorf("failed to parse login page")
	}

	nameHTML, _ := nameElement.Html()
	re := regexp.MustCompile(`Inloggad som: (?P<name>\D*)\d* \[(?P<username>.*)\]`)
	matches := re.FindStringSubmatch(nameHTML)

	name := "N/A"
	username := "N/A"

	if len(matches) >= 3 {
		name = strings.TrimSpace(matches[1])
		username = strings.TrimSpace(matches[2])
	}

	return &LoginInfo{
		Name:     name,
		Username: username,
	}, nil
}
