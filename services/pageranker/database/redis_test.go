package database

import (
	"fmt"
	"testing"
)

func initDb() DataBase {
	db := DataBase{}
	err := db.Connect("localhost:6379", "0", "")
	if err != nil {
		panic(err)
	}
	return db
}

func TestGetPageNodes(t *testing.T) {
	db := initDb()
	pageNodes, err := db.GetPageNodes()
	if err != nil {
		t.Error(err)
	}

	for _, node := range pageNodes {
		fmt.Println(node.Hash, node.OutLinks)
	}
}
