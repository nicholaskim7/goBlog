package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	goBlog "github.com/nicholaskim7/go-blog"
)

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func makeSlug(s string) string {
	// converts user inputed title into url friendly slug to be used as filename
	s = strings.ToLower(s)

	// find any characters that are NOT a-z, 0-9, or hyphen
	reg := regexp.MustCompile(`[^a-z0-9]+`)

	// Replace those invalid characters with a hyphen
	s = reg.ReplaceAllString(s, "-")

	// Trim any leading/trailing hyphens
	s = strings.Trim(s, "-")
	return s
}

func run(args []string, stdout io.Writer) error {
	// Ensure the 'posts' directory exists
	if err := os.MkdirAll("posts", 0755); err != nil {
		return fmt.Errorf("could not create posts directory: %w", err)
	}

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

	// parse write template for form to type up new blog post markdown
	writeTemplate := template.Must(template.ParseFiles("write.gohtml"))
	// register handler for write page
	mux.HandleFunc("GET /write", func(w http.ResponseWriter, r *http.Request) {
		if err := writeTemplate.Execute(w, nil); err != nil {
			log.Printf("error executing write template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	// handler for form submission
	mux.HandleFunc("POST /submit", func(w http.ResponseWriter, r *http.Request) {
		// Parse the form data
		if err := r.ParseForm(); err != nil {
			log.Printf("error parsing form: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		// get the title and markdown content from the form values
		title := r.FormValue("title")
		markdownBody := r.FormValue("markdown")

		// generate url friendly slug
		slug := makeSlug(title)
		postFileName := fmt.Sprintf("posts/%s.md", slug)

		// Write the markdown content to a new file
		err := os.WriteFile(postFileName, []byte(markdownBody), 0644)
		if err != nil {
			log.Printf("could not write file %s: %v", postFileName, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		log.Printf("successfully created post: %s", postFileName)

		// Redirect the user to the new post's page
		http.Redirect(w, r, fmt.Sprintf("/posts/%s", slug), http.StatusFound)
	})

	// Serve static files from the "public" directory at the "/public/" URL path.
	mux.Handle("GET /public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	// Start the web server on port 3030, using our mux to handle requests.
	// This is a blocking call.
	err := http.ListenAndServe(":3030", mux)

	// If the server fails to start, log the error and exit the program.
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
