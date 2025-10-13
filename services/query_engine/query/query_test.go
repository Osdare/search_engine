package query

import (
	"fmt"
	"query_engine/database"
	"testing"
)

func TestGetUrlsFromQuery(t *testing.T) {
	db := database.DataBase{}
	err := db.Connect("localhost:6379", "0", "")
	if err != nil {
		t.Errorf("could not connect to database %v", err)
	}

	result, err := GetRelevantUrls("ooga boogang the osu game", &db, 10)
	if err != nil {
		t.Error(err)
	}

	for _, link := range result {
		fmt.Println(link)
	}
	fmt.Println(len(result))
}
