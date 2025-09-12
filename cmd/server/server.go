package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	goBlog "github.com/nicholaskim7/go-blog"
)

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	// Create a new request router (also called a "serve mux").
	mux := http.NewServeMux()

	postReader := goBlog.FileReader{
		// set up posts directory where our posts can be found
		Dir: "posts",
	}

	// Parse the HTML template file once at startup.
	postTemplate := template.Must(template.ParseFiles("post.gohtml"))

	// Register a handler for GET requests to /posts/{slug}.
	// PostHandler is a factory that creates the actual handler function.
	mux.HandleFunc("GET /posts/{slug}", goBlog.PostHandler(postReader, postTemplate))

	indexTemplate := template.Must(template.ParseFiles("index.gohtml"))
	mux.HandleFunc("GET /", goBlog.IndexHandler(postReader, indexTemplate))

	// Start the web server on port 3030, using our mux to handle requests.
	// This is a blocking call.
	err := http.ListenAndServe(":3030", mux)

	// If the server fails to start, log the error and exit the program.
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
