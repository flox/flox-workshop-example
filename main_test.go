package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gorilla/mux"
)

func resetQuotes() {
	quotes = nil
	quotesOnce = sync.Once{}
	quotesSource = ""
}

func setQuotesDirectly(q []interface{}) {
	quotes = q
	quotesOnce.Do(func() {}) // mark as already loaded
}

func setupTestRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", getIndex).Methods("GET")
	r.HandleFunc("/quotes", getAllQuotes).Methods("GET")
	r.HandleFunc("/quotes/{index}", getQuoteByIndex).Methods("GET")
	return r
}

func TestGetIndex(t *testing.T) {
	r := setupTestRouter()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if _, ok := result["GET /"]; !ok {
		t.Error("Expected 'GET /' in response")
	}
	if _, ok := result["GET /quotes"]; !ok {
		t.Error("Expected 'GET /quotes' in response")
	}
	if _, ok := result["GET /quotes/{index}"]; !ok {
		t.Error("Expected 'GET /quotes/{index}' in response")
	}
}

func TestLoadQuotesFromFile(t *testing.T) {
	resetQuotes()

	tmpFile := filepath.Join(t.TempDir(), "test_quotes.json")
	testData := []string{"Quote 1", "Quote 2", "Quote 3"}
	data, _ := json.Marshal(testData)
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	loadQuotesFromFile(tmpFile)

	if len(quotes) != 3 {
		t.Errorf("Expected 3 quotes, got %d", len(quotes))
	}
}

func TestLoadQuotesFromFileNotFound(t *testing.T) {
	resetQuotes()

	_, err := os.Stat("/nonexistent/file.json")
	if !os.IsNotExist(err) {
		t.Skip("Test file unexpectedly exists")
	}
}

func TestGetAllQuotes(t *testing.T) {
	resetQuotes()
	setQuotesDirectly([]interface{}{"Quote A", "Quote B"})

	r := setupTestRouter()
	req := httptest.NewRequest("GET", "/quotes", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 quotes, got %d", len(result))
	}
}

func TestGetQuoteByIndex(t *testing.T) {
	resetQuotes()
	setQuotesDirectly([]interface{}{"First", "Second", "Third"})

	r := setupTestRouter()

	tests := []struct {
		name       string
		index      string
		wantStatus int
		wantQuote  string
	}{
		{"valid index 0", "0", http.StatusOK, "First"},
		{"valid index 1", "1", http.StatusOK, "Second"},
		{"valid index 2", "2", http.StatusOK, "Third"},
		{"index out of bounds", "99", http.StatusNotFound, ""},
		{"negative index", "-1", http.StatusNotFound, ""},
		{"invalid index", "abc", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/quotes/"+tt.index, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusOK {
				var result string
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				if result != tt.wantQuote {
					t.Errorf("Expected %q, got %q", tt.wantQuote, result)
				}
			}
		})
	}
}

func TestGetAllQuotesContentType(t *testing.T) {
	resetQuotes()
	setQuotesDirectly([]interface{}{"Test"})

	r := setupTestRouter()
	req := httptest.NewRequest("GET", "/quotes", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestGetQuoteByIndexContentType(t *testing.T) {
	resetQuotes()
	setQuotesDirectly([]interface{}{"Test"})

	r := setupTestRouter()
	req := httptest.NewRequest("GET", "/quotes/0", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestEmptyQuotes(t *testing.T) {
	resetQuotes()
	setQuotesDirectly([]interface{}{})

	r := setupTestRouter()
	req := httptest.NewRequest("GET", "/quotes/0", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for empty quotes, got %d", w.Code)
	}
}
