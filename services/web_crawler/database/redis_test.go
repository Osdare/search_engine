package database

import (
	"reflect"
	"testing"
	"utils"
	"web_crawler/types"
)

func TestAddDomainGetDomain(t *testing.T) {
	db := DataBase{}
	err := db.Connect("localhost:6379", "0", "")
	if err != nil {
		t.Errorf("could not connect to database %v", err)
	}

	domain := types.Domain{
		Name:        "google.com",
		CrawlDelay:  10,
		LastCrawled: utils.GetTimeInt(),
		Allowed:     []string{},
		Disallowed:  []string{},
	}

	err = db.AddDomain(domain)
	if err != nil {
		t.Errorf("domain: %v could not be added to database %v", domain.Name, err)
	}

	dbDomain, err := db.GetDomain(domain.Name)
	if err != nil {
		t.Errorf("got en error when trying to fetch domain %v, %v", domain.Name, err)
	}

	if !reflect.DeepEqual(dbDomain, domain) {
		t.Error("domains are not equal ", domain, dbDomain)
	}
}

func TestUrlQueue(t *testing.T) {
	tesurls := []string{
		"google.com/oogabooga",
		"x.com/nooo",
		"instagram.com/shit",
		"instagram.com/shit",
		"instagram.com/shit",
		"example.ax/fisk",
	}

	db := DataBase{}
	err := db.Connect("localhost:6379", "0", "")
	if err != nil {
		t.Errorf("could not connect to db %v", err)
	}

	for _, url := range tesurls {
		err = db.PushUrl(url)
		if err != nil {
			t.Errorf("could not push url %v %v", url, err)
		}

	}

	length, err := db.UrlQueueLength()
	if err != nil {
		t.Error(err)
	}

	if length != 4 {
		t.Errorf("expected 4 got %d", length)
	}

	//ADD a url to the db and then try to re-add it to the queue

	url, err := db.PopUrl()
	if err != nil {
		t.Error(err)
	}

	urlPage := types.Page{
		NormUrl: url,
		Content: "i am a random piece of html",
	}

	err = db.AddPage(urlPage)
	if err != nil {
		t.Error(err)
	}

	err = db.PushUrl(url)
	if err != nil {
		t.Error(err)
	}

	length, err = db.UrlQueueLength()
	if err != nil {
		t.Error(err)
	}

	if length != 3 {
		t.Errorf("length was %d expected 3", length)
	}

	db.client.FlushAll(db.ctx)
}
