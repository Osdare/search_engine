package query

import (
	"fmt"
	"query_engine/database"
	"testing"
)

func TestGetImages(t *testing.T) {
	db := database.DataBase{}
	err := db.Connect("localhost:6379", "0", "")
	if err != nil {
		t.Error(err)
	}

	query := "osu gamer"
	imagelinks, err := GetImages(&db, query)
	if err != nil {
		t.Error(err)
	}

	for _, link := range imagelinks {
		fmt.Println(link)
	}
}
