package database

import (
	"fmt"
	"testing"
)

func TestDb(t *testing.T) {
	db := DataBase{}

	err := db.Connect("localhost:6379", "0", "")
	if err != nil {
		t.Errorf("could not connect to database %v", err)
	}
	
	ten, err := db.GetTopXIndices("enter", 10) 
	if err != nil {
		t.Error(err)
	}
	fmt.Println(ten)
}