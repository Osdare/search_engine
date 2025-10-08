package main

import (
	"indexer/database"
	pagerank "indexer/page_rank"
)

func main() {
	db := database.DataBase{}
	err := db.Connect("localhost:6379", "0", "")
	if err != nil {
		panic(err)
	}

	pageNodes, err := db.GetPageNodes()
	if err != nil {
		panic(err)
	}

	pageRanks := pagerank.GetPageRanks(pageNodes)

	err = db.AddPageRanks(pageRanks)
	if err != nil {
		panic(err)
	}
}
