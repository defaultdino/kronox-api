package parsers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/defaultdino/kronox-api/pkg/models"
)

func (s *service) ParseProgrammes(html string) ([]*models.Programme, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var programmes []*models.Programme

	doc.Find("a[target='_blank']").Slice(2, goquery.ToEnd).Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		re := regexp.MustCompile(`resurser=(.*)$`)
		matches := re.FindStringSubmatch(href)
		if len(matches) < 2 {
			return
		}

		id := matches[1]
		text := strings.TrimSpace(s.Text())
		parts := strings.Split(text, ",")

		if len(parts) < 2 {
			return
		}

		title := strings.TrimSpace(parts[0])
		subtitle := removeDuplicateWords(strings.TrimSpace(parts[1]))

		programmes = append(programmes, &models.Programme{
			Title:    title,
			Subtitle: strings.TrimSpace(subtitle),
			Id:       id,
		})
	})

	return programmes, nil
}
