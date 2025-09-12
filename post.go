package goBlog

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

// query all the meta data for all blog posts
type MetadataQuerier interface {
	Query() ([]PostMetadata, error)
}

// meta data at the top of each blog post
type PostMetadata struct {
	Slug        string
	Title       string    `toml:"title"`
	Author      Author    `toml:"author"`
	Description string    `toml:"description"`
	Date        time.Time `toml:"date"`
}

// SlugReader defines a contract for any type that can read content based on a slug.
// Anything with a Read(string) (string, error) method can be treated as a SlugReader.
type SlugReader interface {
	Read(slug string) (string, error)
}

// FileReader is a type that implements the SlugReader interface by reading from local files.
type FileReader struct {
	// Directory to find blog posts in
	Dir string
}

// Read finds and reads a markdown file from the disk corresponding to the slug.
// FileReader has a method named Read with same signature. FileReader is recognized as implementing SlugReader
func (fr FileReader) Read(slug string) (string, error) {
	// include directory to path
	slugPath := filepath.Join(fr.Dir, slug+".md")
	// open mark down file
	f, err := os.Open(slugPath)
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

func (fr FileReader) Query() ([]PostMetadata, error) {
	// include directory to path
	postsPath := filepath.Join(fr.Dir, "*.md")
	// Glob all the markdown files in the current directory
	filenames, err := filepath.Glob(postsPath)
	if err != nil {
		return nil, fmt.Errorf("querying for files: %w", err)
	}
	var posts []PostMetadata
	for _, filename := range filenames {
		// Open each file
		f, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("opening file %q: %w", filename, err)
		}
		defer f.Close()
		var post PostMetadata
		//parse frontmatter into post
		_, err = frontmatter.Parse(f, &post)
		if err != nil {
			return nil, fmt.Errorf("parsing frontmatter for file %s: %w", filename, err)
		}
		// get the base filename and remove the .md suffix to get the slug
		post.Slug = strings.TrimSuffix(filepath.Base(filename), ".md")
		// append post to posts slice
		posts = append(posts, post)
		//TODO Open file, Read metadata, place into posts slice
	}
	return posts, nil
}

type PostData struct {
	Content template.HTML
	Author  Author `toml:"author"`
	Title   string `toml:"title"`
}

type Author struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
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

		// extract frontmatter from postMarkDown
		var post PostData
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
		// put remaining content into post
		post.Content = template.HTML(buf.String())

		err = tpl.Execute(w, post)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}

type IndexData struct {
	Posts []PostMetadata
}

func IndexHandler(mq MetadataQuerier, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := mq.Query()
		if err != nil {
			http.Error(w, "Error querying posts", http.StatusInternalServerError)
			return
		}
		data := IndexData{
			Posts: posts,
		}
		err = tpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}
