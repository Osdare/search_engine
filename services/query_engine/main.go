package main

import (
	"fmt"
	"log"
	"net/http"
	"query_engine/database"
	"query_engine/query"
)

const allowedOrigin = "http://127.0.0.1:5500"

func main() {

	db := database.DataBase{}
	err := db.Connect("localhost:6379", "0", "")
	if err != nil {
		panic(err)
	}

	queryHandler := func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin != allowedOrigin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		message := r.FormValue("message")
		if message == "" {
			message = "You didn't send a message!"
		}

		links, err := query.GetRelevantUrls(message, &db, 10)
		if err != nil {
			http.Error(w, "Error while handling query", http.StatusInternalServerError)
		}

		var linkString string
		for _, link := range links {
			linkString += link + " "
		}

		fmt.Println("Received query:", message)
		fmt.Fprintf(w, "%s", linkString)
	}

	http.HandleFunc("/echo", queryHandler)
	port := ":8080"
	fmt.Printf("Server running at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
