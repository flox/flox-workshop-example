package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var (
	quotes       []interface{}
	quotesOnce   sync.Once
	quotesSource string
)

func loadQuotes() {
	if quotesSource == "redis" {
		loadQuotesFromRedis()
	} else {
		loadQuotesFromFile(quotesSource)
	}
}

func loadQuotesFromFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read quotes file: %v", err)
	}
	if err := json.Unmarshal(data, &quotes); err != nil {
		log.Fatalf("Failed to parse quotes from file: %v", err)
	}
	fmt.Printf("Loaded quotes from %s\n", path)
}

func loadQuotesFromRedis() {
	redisHost := "localhost"
	redisPort := os.Getenv("REDISPORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	defer redisClient.Close()

	retries := 3
	var data string
	var err error

	for i := 0; i < retries; i++ {
		data, err = redisClient.Get(ctx, "quotesjson").Result()
		if err == nil {
			break
		}
		log.Printf("Retry %d: Failed to fetch quotes from Redis: %v", i+1, err)
	}

	if err != nil {
		log.Fatalf("All retries failed: %v", err)
	}

	if err := json.Unmarshal([]byte(data), &quotes); err != nil {
		log.Fatalf("Failed to parse quotes data: %v", err)
	}

	fmt.Println("Loaded quotes from Redis")
}

func ensureQuotesLoaded() {
	quotesOnce.Do(loadQuotes)
}

func getAllQuotes(w http.ResponseWriter, r *http.Request) {
	ensureQuotesLoaded()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotes)
}

func getQuoteByIndex(w http.ResponseWriter, r *http.Request) {
	ensureQuotesLoaded()
	vars := mux.Vars(r)
	indexStr, ok := vars["index"]
	if !ok {
		http.Error(w, `{"error":"Index not provided"}`, http.StatusBadRequest)
		return
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(quotes) {
		http.Error(w, `{"error":"Index out of bounds"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotes[index])
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: quotes-app <quotes.json | redis>")
		fmt.Fprintln(os.Stderr, "  quotes.json  - path to a JSON file containing quotes")
		fmt.Fprintln(os.Stderr, "  redis        - load quotes from Redis")
		os.Exit(1)
	}

	quotesSource = os.Args[1]
	if quotesSource != "redis" {
		if _, err := os.Stat(quotesSource); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: file %q does not exist\n", quotesSource)
			os.Exit(1)
		}
	}

	r := mux.NewRouter()
	r.HandleFunc("/quotes", getAllQuotes).Methods("GET")
	r.HandleFunc("/quotes/{index}", getQuoteByIndex).Methods("GET")

	addr := ":3000"
	fmt.Printf("Server running on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
