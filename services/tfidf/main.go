package main

import (
	"tfidf/database"
	"utils"
)

func main() {
	redisHost := utils.GetEnv("REDIS_HOST", "localhost")
	redisPort := utils.GetEnv("REDIS_PORT", "6379")
	redisPassword := utils.GetEnv("REDIS_PASSWORD", "")
	redisDB := utils.GetEnv("REDIS_DB", "0")

	db := database.DataBase{}
	if err := db.Connect(redisHost+":"+redisPort, redisDB, redisPassword); err != nil {
		panic(err)
	}

	err := db.StreamIndices(1000)
	if err != nil {
		panic(err)
	}
}
