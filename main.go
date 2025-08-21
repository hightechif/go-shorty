package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	_ "github.com/gogo/status"
)

// =======================================================================================
// URLStore Module - This is the core logic from our previous project.
// It is now thread-safe using a sync.RWMutex to handle concurrent web requests.
// =======================================================================================

type URLStore struct {
	urls     map[string]string
	mu       sync.RWMutex // Mutex to make our map safe for concurrent access
	filename string
}

func (s *URLStore) Add(longURL string, customKey *string) (string, error) {
	var shortKey string

	if customKey != nil {
		shortKey = *customKey
	} else {
		keyBytes := make([]byte, 4)
		if _, err := rand.Read(keyBytes); err != nil {
			return "", err
		}
		shortKey = hex.EncodeToString(keyBytes)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[shortKey] = longURL

	go func() {
		if err := s.save(); err != nil {
			log.Printf("Error saving to file: %v", err)
		}
	}()

	return shortKey, nil
}

func (s *URLStore) Get(shortKey string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	longURL, found := s.urls[shortKey]
	return longURL, found
}

func (s *URLStore) save() error {
	data, err := json.Marshal(s.urls)
	if err != nil {
		return err
	}
	return os.WriteFile(s.filename, data, 0644)
}

func (s *URLStore) load() error {
	data, err := os.ReadFile(s.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Unmarshal(data, &s.urls)
}

// =======================================================================================
// Web Server Logic - This is the new part that exposes our URLStore via HTTP.
// =======================================================================================

type urlHandler struct {
	store *URLStore
}

func (h *urlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/shorty" {
		if r.Method == http.MethodPost {
			h.handlePost(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if r.Method == http.MethodGet {
		h.handleGet(w, r)
		return
	}

	http.NotFound(w, r)
}

func (h *urlHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	shortKey := r.URL.Path[1:]
	if shortKey == "" {
		http.Error(w, "Welcome to Go-Shorty! Use POST to /shorty to create a short URL.", http.StatusOK)
		return
	}

	longURL, found := h.store.Get(shortKey)
	if !found {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

func (h *urlHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		URL       string  `json:"url"`
		CustomKey *string `json:"customKey,omitempty"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if requestData.URL == "" {
		http.Error(w, "URL field is required", http.StatusBadRequest)
		return
	}

	shortKey, err := h.store.Add(requestData.URL, requestData.CustomKey)
	if err != nil {
		http.Error(w, "Failed to create short key", http.StatusInternalServerError)
		return
	}

	responseData := struct {
		ShortKey string `json:"shortKey"`
	}{
		ShortKey: shortKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseData)
}

func NewURLStore(filename string) *URLStore {
	store := &URLStore{
		urls:     make(map[string]string),
		filename: filename,
	}
	if err := store.load(); err != nil {
		log.Printf("Warning: could not load data from %s: %v", filename, err)
	}
	return store
}

func main() {
	const filename = "urls.json"
	store := NewURLStore(filename)
	handler := &urlHandler{store: store}

	fmt.Println("Starting Go-Shorty URL shortener API on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
