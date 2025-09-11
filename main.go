package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

func main() {
	// Create a new request router (also called a "serve mux").
	mux := http.NewServeMux()

	// Parse the HTML template file once at startup.
	postTemplate := template.Must(template.ParseFiles("post.gohtml"))

	// Register a handler for GET requests to /posts/{slug}.
	// PostHandler is a factory that creates the actual handler function.
	mux.HandleFunc("GET /posts/{slug}", PostHandler(FileReader{}, postTemplate))

	// Start the web server on port 3030, using our mux to handle requests.
	// This is a blocking call.
	err := http.ListenAndServe(":3030", mux)

	// If the server fails to start, log the error and exit the program.
	if err != nil {
		log.Fatal(err)
	}
}

// SlugReader defines a contract for any type that can read content based on a slug.
// Anything with a Read(string) (string, error) method can be treated as a SlugReader.
type SlugReader interface {
	Read(slug string) (string, error)
}

// FileReader is a type that implements the SlugReader interface by reading from local files.
type FileReader struct{}

// Read finds and reads a markdown file from the disk corresponding to the slug.
// FileReader has a method named Read with same signature. FileReader is recognized as implementing SlugReader
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
} // FileReader can now be used anywhere a SlugReader is expected.

type PostData struct {
	Content template.HTML
	Author  string
	Title   string
}

// PostHandler is a "higher-order function". It takes a SlugReader dependency
// and returns an http.HandlerFunc that uses it.
func PostHandler(sl SlugReader, tpl *template.Template) http.HandlerFunc {
	mdRenderer := goldmark.New(
		goldmark.WithExtensions(
			// syntax highlighting
			highlighting.NewHighlighting(
				// change theme of syntax highlighting
				highlighting.WithStyle("dracula"),
			),
		),
	)

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

		var buf bytes.Buffer
		// convert the postMarkDown into a buffer
		// The mdRenderer.Convert method requires a byte slice, (interface returns string so convert to byte slice)
		err = mdRenderer.Convert([]byte(postMarkdown), &buf)
		if err != nil {
			panic(err)
		}

		err = tpl.Execute(w, PostData{
			// Cast to template.HTML to prevent Go's template engine from auto-escaping the HTML.
			Content: template.HTML(buf.String()),
			Author:  "Nicholas Kim",
			Title:   "My Blog", // hardcoded for now
		})
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}
