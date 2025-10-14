package main

import (
	"fmt"
	"log"
	"net/http"
	"query_engine/database"
	"query_engine/query"
)

const (
	allowedOrigin = "http://127.0.0.1:5500"
	serverPort    = ":8080"
)

func main() {
	db := &database.DataBase{}
	if err := db.Connect("localhost:6379", "0", ""); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis")

	http.HandleFunc("/links", withCORS(makeHandler(db, query.GetRelevantUrls)))
	http.HandleFunc("/images", withCORS(makeHandler(db, query.GetImages)))

	fmt.Printf("Server running at http://localhost%s\n", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, nil))
}

func makeHandler(db *database.DataBase, handlerFunc interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setupCORS(w, r)
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
			http.Error(w, "Missing 'message' parameter", http.StatusBadRequest)
			return
		}

		var (
			links []string
			err   error
		)

		switch f := handlerFunc.(type) {
		case func(string, *database.DataBase, int) ([]string, error):
			links, err = f(message, db, 20)
		case func(*database.DataBase, string) ([]string, error):
			links, err = f(db, message)
		default:
			http.Error(w, "Invalid handler function", http.StatusInternalServerError)
			return
		}

		if err != nil {
			log.Printf("query error: %v", err)
			http.Error(w, "Error while handling query", http.StatusInternalServerError)
			return
		}

		fmt.Println("Received query:", message)
		fmt.Fprint(w, joinLinks(links))
	}
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if origin != "" && origin != allowedOrigin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

func setupCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != allowedOrigin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func joinLinks(links []string) string {
	result := ""
	for i, link := range links {
		if i > 0 {
			result += " "
		}
		result += link
	}
	return result
}
