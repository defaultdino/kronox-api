package kronox

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/defaultdino/kronox-api/pkg/models"
)

var programmeResurserRE = regexp.MustCompile(`resurser=(.*)$`)

func ParseProgrammes(html string) ([]*models.Programme, error) {
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
		matches := programmeResurserRE.FindStringSubmatch(href)
		if len(matches) < 2 {
			return
		}
		id := matches[1]

		parts := strings.SplitN(strings.TrimSpace(s.Text()), ",", 2)
		if len(parts) < 2 {
			return
		}
		title := strings.TrimSpace(parts[0])
		subtitle := strings.TrimSpace(removeDuplicateWords(strings.TrimSpace(parts[1])))

		programmes = append(programmes, &models.Programme{
			Id:       id,
			Title:    title,
			Subtitle: subtitle,
		})
	})

	return programmes, nil
}
