package crawler

import (
	"fmt"
	"testing"
	"web_crawler/database"
	"web_crawler/parser"
	"web_crawler/utilities"
)

func TestCrawlAndParse(t *testing.T) {
	url := "https://osu.ppy.sh/users/5070783"
	html, _, err := Crawl(url)
	if err != nil {
		t.Error(err)
	}

	_, urls, _, _ := parser.ParseBody(url, html)

	fmt.Println(len(urls))
	normUrls, err := utilities.NormalizeUrlSlice(url, urls)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(len(normUrls))
}

func TestCrawlJob(t *testing.T) {
	db := database.DataBase{}
	db.Connect("localhost:6379", "0", "")

	err := db.PushUrl("https://en.wikipedia.org/wiki/Osu!")
	if err != nil {
		t.Error(err)
	}

	CrawlJob(&db)
}
