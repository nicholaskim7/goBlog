package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/nicholaskim7/go-blog/internal/handler"
	"github.com/nicholaskim7/go-blog/internal/post"
)

func main() {
	// call run function
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Initialize Dependencies
	postReader := post.FileReader{
		// directory to find blog posts in
		Dir: "posts",
	}

	// Register routes
	mux := handler.RegisterRoutes(postReader)

	// Start the HTTP server on port 3030
	log.Println("Starting server on :3030")
	err := http.ListenAndServe(":3030", mux)
	if err != nil {
		return fmt.Errorf("could not start server: %w", err)
	}
	return nil
}
