package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	// Create a new request router (also called a "serve mux").
	mux := http.NewServeMux()

	// Register a handler for GET requests to /posts/{slug}.
	// PostHandler is a factory that creates the actual handler function.
	mux.HandleFunc("GET /posts/{slug}", PostHandler(FileReader{}))

	// Start the web server on port 3030, using our mux to handle requests.
	// This is a blocking call.
	err := http.ListenAndServe(":3030", mux)

	// If the server fails to start, log the error and exit the program.
	if err != nil {
		log.Fatal(err)
	}
}

// SlugReader defines a contract for any type that can read content based on a slug.
// This use of an interface allows us to swap out the implementation (e.g., for testing).
type SlugReader interface {
	Read(slug string) (string, error)
}

// FileReader is a type that implements the SlugReader interface by reading from local files.
type FileReader struct{}

// Read finds and reads a markdown file from the disk corresponding to the slug.
func (fr FileReader) Read(slug string) (string, error) {
	// open mark down file
	f, err := os.Open(slug + ".md")
	if err != nil {
		return "", err
	}
	defer f.Close()
	// read all contents of mark down file
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	// cast contents to string
	return string(b), nil
}

// PostHandler is a "higher-order function". It takes a SlugReader dependency
// and returns an http.HandlerFunc that uses it.
func PostHandler(sl SlugReader) http.HandlerFunc {
	// This returned function is the actual handler that will process HTTP requests.
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the 'slug' value from the request URL path
		slug := r.PathValue("slug")

		// Use the provided SlugReader to get the post content.
		postMarkdown, err := sl.Read(slug)
		if err != nil {
			//Todo handle different errors in the future
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		// write contents of markdown file to response writer
		fmt.Fprint(w, postMarkdown)
	}
}
