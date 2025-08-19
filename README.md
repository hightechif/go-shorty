# Go-Shorty: A Simple URL Shortener API

A lightweight, persistent URL shortener built with Go's standard library. This project provides a simple REST API to create short links for long URLs and redirects users when they access the short link.

## ✨ Features

- **Create Short URLs**: Generate a unique, random 8-character key for any long URL.
- **Redirect Service**: Automatically redirects users from the short URL to the original destination.
- **Data Persistence**: URL mappings are saved to a local urls.json file, so data is not lost on restart.
- **Concurrent Ready**: Uses a mutex to safely handle multiple simultaneous requests.
- **Minimalist**: Built entirely with the Go standard library, no external dependencies needed.

## 🚀 API Endpoints

The API is simple and consists of two main endpoints.

1. **Create a Short URL**

   Creates a new short URL for a given long URL.

   - **Endpoint**: /shorty
   - **Method**: POST
   - **Request Body (JSON):**
     ```json
     {
       "url": "https://www.google.com/search?q=golang+projects"
     }
     ```
   - **Success Response (201 Created):**
     ```json
     {
       "short_key": "a1b2c3d4"
     }
     ```
   - **Request Body (JSON) with customKey:**
     ```json
     {
       "url": "https://www.google.com/search?q=golang+projects",
       "customKey": "myurl"
     }
     ```
   - **Success Response (201 Created):**
     ```json
     {
       "short_key": "myurl"
     }
     ```

2. Redirect to Original URL

   Redirects the user to the original long URL associated with a short key.

   - **Endpoint:** `/{shortKey}`
   - **Method:** `GET`
   - **Example Path:** `/a1b2c3d4`
   - **Success Response** `(302 Found)`:
     - The server responds with an HTTP redirect to the original URL.

## 🛠️ Getting Started

Follow these instructions to get a copy of the project up and running on your local machine.

**Prerequisites**

- Go (version 1.18 or higher recommended)

**Installation & Running**

1. **Clone the repository (or save the code):**

   If you have a repository, clone it. Otherwise, save the code into a file named main.go.

   ```
   git clone https://github.com/your-username/go-shorty.git
   cd go-shorty
   ```

2. Run the server:

   Execute the go run command in your terminal. The server will start on port 8080.

   ```
   go run main.go
   ```

   You should see the following output:

   ```
   Starting Go-Shorty URL shortener API on :8080
   ```

## 📋 Usage Example

You can interact with the API using a tool like curl or by visiting the endpoints in your browser.

1. Create a short link:

   ```
   curl -X POST -H "Content-Type: application/json" \
   -d '{"url":"https://en.wikipedia.org/wiki/Go_(programming_language)"}' \
   http://localhost:8080/shorty
   ```

   This will return a short key, for example: {"short_key":"e5f6a7b8"}

2. Use the short link:

   Open your web browser and navigate to the following URL, replacing e5f6a7b8 with the key you received:

   ```
   http://localhost:8080/e5f6a7b8
   ```

   You will be automatically redirected to the Wikipedia page for Go.

**For shorten url with custom key**

1. Create a short link with custom key:

   ```
   curl -X POST -H "Content-Type: application/json" \
   -d '{"url":"https://en.wikipedia.org/wiki/Go_(programming_language)" \
   "customKey":"myurl"}' \
   http://localhost:8080/shorty
   ```

   This will return a short key, for example: {"short_key":"myurl"}

2. Use your custom short link:

   Open your web browser and navigate to the following URL, replacing myurl with the key you received:

   ```
   http://localhost:8080/myurl
   ```

   You will be automatically redirected to the Wikipedia page for Go.

## 📁 File Structure

The project is contained within a single file for simplicity.

```
.
└── main.go # All the application logic
└── urls.json # The data file (created automatically)
```

## 🏗️ Built With

- **Go** - The programming language
- **Standard Library Packages:**
  - `net/http` - For the web server
  - `encoding/json` - For handling JSON data
  - `os & io` - For file I/O
  - `sync` - For concurrency control (Mutex)
  - `crypto/rand` - For generating random keys
