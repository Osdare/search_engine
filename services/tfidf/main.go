package main

import (
	"tfidf/database"
)

func main() {
	db := database.DataBase{}	
	err := db.Connect("localhost:6379", "0", "")
	if err != nil {
		panic(err)
	}

	err = db.StreamIndices(1000)
	if err != nil {
		panic(err)
	}
}