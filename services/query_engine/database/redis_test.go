package database

import (
	"fmt"
	"testing"
)

func TestGetCandidateLinks(t *testing.T) {
	db := DataBase{}

	err := db.Connect("localhost:6379", "0", "")
	if err != nil {
		t.Errorf("could not connect to database %v", err)
	}

	words := []string{"osu"}	

	candidateLinks, err := db.GetCandidateLinks(words, 100)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(candidateLinks)
}
