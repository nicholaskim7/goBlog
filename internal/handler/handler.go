package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/nicholaskim7/go-blog/internal/post"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

// takes in the MetadataQuerier interface and queries all the metadata for all blog posts
// look at FileReader Query method to see how it implements MetadataQuerier interface
func IndexHandler(mq post.MetadataQuerier, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := mq.Query()
		if err != nil {
			http.Error(w, "Error querying posts", http.StatusInternalServerError)
			return
		}
		data := post.IndexData{
			Posts: posts,
		}
		err = tpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}

// PostHandler is a "higher-order function". It takes a SlugReader dependency
// and returns an http.HandlerFunc that uses it.
func PostHandler(sl post.SlugReader, tpl *template.Template) http.HandlerFunc {
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
		// look at FileReader Read method to see how it implements SlugReader interface
		postMarkdown, err := sl.Read(slug)
		if err != nil {
			//Todo handle different errors in the future
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		var post post.PostData
		// Reads the frontmatter metadata from postMarkDown and fills out the post struct
		remainingMd, err := frontmatter.Parse(strings.NewReader(postMarkdown), &post)
		if err != nil {
			http.Error(w, "Error parsing front matter", http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		// convert the remaining postMarkDown into a buffer
		// The mdRenderer.Convert method requires a byte slice, (interface returns string so convert to byte slice)
		err = mdRenderer.Convert([]byte(remainingMd), &buf)
		if err != nil {
			panic(err)
		}
		// put remaining postmarkdown content into post content
		post.Content = template.HTML(buf.String())

		// Render the post using the provided template
		err = tpl.Execute(w, post)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}

// register handler for write page
func WriteHandler(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// render the write template which is a form for submitting new blog posts
		if err := tpl.Execute(w, nil); err != nil {
			log.Printf("error executing write template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// handler for form submission
func SubmitHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the form data
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Printf("error parsing form: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		// get the title, markdown content from the form values
		title := r.FormValue("title")
		markdownBody := r.FormValue("markdown")

		// get file headers from the multipartform
		imageFiles := r.MultipartForm.File["imageUpload"]

		// loop through the files and save them to pictures directory
		for _, imgFile := range imageFiles {
			file, err := imgFile.Open()
			if err != nil {
				http.Error(w, "Unable to open uploaded file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// create destination if needed
			dstPath := filepath.Join("public", "pictures", imgFile.Filename)
			dst, err := os.Create(dstPath)
			if err != nil {
				http.Error(w, "Unable to create destination file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			// Copy the content
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Unable to save file", http.StatusInternalServerError)
				return
			}
		}

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
