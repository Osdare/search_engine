package handlers

import (
	"fmt"
	"testing"
	"web_crawler/types"
)

func TestRobotsToDomain(t *testing.T) {

	domainName := "https://osu.ppy.sh"

	domain, err := GetRobotsFromDomain(domainName)
	if err != nil {
		t.Error(err)
	}

	if len(domain.Disallowed) > 0 || len(domain.Allowed) > 0 {
		t.Errorf("Expected no robots")
	}

	fmt.Println(domain)

}

func TestCanCrawl(t *testing.T) {
	domainG := types.Domain{
		Name:        "https://google.com",
		CrawlDelay:  1,
		LastCrawled: 1234,
		Allowed:     []string{},
		Disallowed:  []string{},
	}

	expected := map[string]bool{
		"https://google.com/search/about":                     true,
		"https://google.com/search/foo":                       false,
		"https://google.com/search/howsearchworks":            true,
		"https://google.com/search/yeah":                      false,
		"https://google.com/example/page/":                    true,
		"https://google.com/example/page/disallowed.gif":      false,
		"https://google.com/example/page/asdf.gif?size=large": true,
		"https://google.com/example/allowed.gif":              true,
		"https://google.com/lolo.gif":                         false,
	}

	for key := range expected {
		res, _, err := CanCrawl(key, domainG)
		if err != nil {
			t.Error(err)
		}
		if !res {
			t.Errorf("expected true got: %v", res)
		}
	}
}
