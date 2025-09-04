package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"context"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var quotes []interface{}

func loadQuotes() {
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

func getAllQuotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotes)
}

func getQuoteByIndex(w http.ResponseWriter, r *http.Request) {
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
	loadQuotes()

	r := mux.NewRouter()
	r.HandleFunc("/quotes", getAllQuotes).Methods("GET")
	r.HandleFunc("/quotes/{index}", getQuoteByIndex).Methods("GET")

	addr := ":3000"
	fmt.Printf("Server running on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
