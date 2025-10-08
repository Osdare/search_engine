package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"web_crawler/crawler"
	"web_crawler/database"
)

// asdfasdfasdf
func main() {
	fmt.Println(":)")

	//env
	godotenv.Load("useragent.env")
	fmt.Println(os.Getenv("USER_AGENT"))

	db := database.DataBase{}
	db.Connect("localhost:6379", "0", "")

	//seed urls should be different urls preferably as many as the amount of crawler workers
	seeds := []string{"https://en.wikipedia.org/wiki/Osu!", "https://osu.ppy.sh/", "https://www.wikihow.com/Play-osu!", "https://github.com/ppy/osu", "https://www.osu.edu/"}
	for _, seed := range seeds {
		err := db.PushUrl(seed)
		if err != nil {
			panic(err)
		}
	}

	//CTRL+C stops the program properly
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("\nStopping...")
		cancel()
	}()

	numWorkers := 2
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := range numWorkers {
		go func(workerID int) {
			defer wg.Done()

			jitter := time.Duration(rand.Intn(3000)) * time.Millisecond
			time.Sleep(jitter)

			fmt.Printf("Worker %d started after %v\n", workerID, jitter)
			crawler.Start(ctx, &db)
		}(i)
	}

	wg.Wait()
	fmt.Println("All workers have stopped.")
}
