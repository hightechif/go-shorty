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

// URLStore holds the mapping from short keys to original URLs.
type URLStore struct {
	urls     map[string]string
	mu       sync.RWMutex // Mutex to make our map safe for concurrent access
	filename string
}

// Add generates a short key for a long URL and stores it.
func (s *URLStore) Add(longURL string, customKey *string) (string, error) {
	var shortKey string

	// 1. Decide which key to use first.
	if customKey != nil {
		// Use the provided custom key.
		shortKey = *customKey
	} else {
		// No custom key, so generate a random one.
		keyBytes := make([]byte, 4)
		if _, err := rand.Read(keyBytes); err != nil {
			return "", err // Failed to generate key
		}
		shortKey = hex.EncodeToString(keyBytes)
	}

	// Lock the map for writing to prevent race conditions.
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[shortKey] = longURL

	// Asynchronously save to the file so we don't block the HTTP response.
	go func() {
		if err := s.save(); err != nil {
			log.Printf("Error saving to file: %v", err)
		}
	}()

	return shortKey, nil
}

// Get retrieves the long URL for a given short key.
func (s *URLStore) Get(shortKey string) (string, bool) {
	// Use a Read Lock for getting data, allowing multiple readers at once.
	s.mu.RLock()
	defer s.mu.RUnlock()
	longURL, found := s.urls[shortKey]
	return longURL, found
}

// save writes the URL map to a file in JSON format.
func (s *URLStore) save() error {
	data, err := json.Marshal(s.urls)
	if err != nil {
		return err
	}
	return os.WriteFile(s.filename, data, 0644)
}

// load reads from a JSON file and populates the URL map.
func (s *URLStore) load() error {
	data, err := os.ReadFile(s.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, which is fine.
		}
		return err
	}
	// Before loading, lock for writing.
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Unmarshal(data, &s.urls)
}

// =======================================================================================
// Web Server Logic - This is the new part that exposes our URLStore via HTTP.
// =======================================================================================

// urlHandler is our main HTTP handler. It holds a reference to the URLStore.
type urlHandler struct {
	store *URLStore
}

// ServeHTTP is the main router. It directs traffic based on the HTTP method and URL path.
func (h *urlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Route for creating a new short URL: POST /shorty
	if r.URL.Path == "/shorty" {
		if r.Method == http.MethodPost {
			h.handlePost(w, r)
		} else {
			// Any other method to /shorty is not allowed
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// Route for redirecting using a short key: GET /{shortKey}
	if r.Method == http.MethodGet {
		h.handleGet(w, r)
		return
	}

	// If no route matches, return a 404 Not Found error.
	http.NotFound(w, r)
}

// handleGet retrieves a URL and redirects the client.
func (h *urlHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	// The short key is the path itself, removing the leading "/"
	shortKey := r.URL.Path[1:]
	if shortKey == "" {
		// Updated welcome message
		http.Error(w, "Welcome to Go-Shorty! Use POST to /shorty to create a short URL.", http.StatusOK)
		return
	}

	longURL, found := h.store.Get(shortKey)
	if !found {
		http.NotFound(w, r)
		return
	}

	// Redirect the client to the original URL.
	http.Redirect(w, r, longURL, http.StatusFound)
}

// handlePost creates a new short URL.
func (h *urlHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	// Define a struct to decode the incoming JSON request.
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

	// Define a struct for the JSON response.
	responseData := struct {
		ShortKey string `json:"shortKey"`
	}{
		ShortKey: shortKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseData)
}

// NewURLStore creates and returns a new URLStore, loading data from a file.
func NewURLStore(filename string) *URLStore {
	store := &URLStore{
		urls:     make(map[string]string),
		filename: filename,
	}
	// Load existing data, but don't block startup if it fails.
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
	// The http.ListenAndServe function starts the server.
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
