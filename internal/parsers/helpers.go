package parsers

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func isSessionExpired(doc *goquery.Document) bool {
	if doc.Find("#boka-dialog-login").Length() > 0 {
		return true
	}

	if strings.Contains(doc.Text(), "Din användare har inte rättigheter att skapa resursbokningar") {
		return true
	}

	if strings.Contains(doc.Text(), "session") && strings.Contains(doc.Text(), "expired") {
		return true
	}

	return false
}
