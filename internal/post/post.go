package post

import (
	"html/template"
	"time"
)

// contract for querying all the meta data for all blog posts
type MetadataQuerier interface {
	Query() ([]PostMetadata, error)
}

// SlugReader defines a contract for any type that can read content based on a slug.
// Anything with a Read(string) (string, error) method can be treated as a SlugReader.
type SlugReader interface {
	Read(slug string) (string, error)
}

// meta data at the top of each blog post
type PostMetadata struct {
	Slug        string
	Title       string    `toml:"title"`
	Author      Author    `toml:"author"`
	Description string    `toml:"description"`
	Date        time.Time `toml:"date"`
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

type IndexData struct {
	Posts []PostMetadata
}
