package post

import (
	"html/template"
	"time"
)

// contract for querying all the metadata for all blog posts
// Any type that wants to be considered a MetadataQuerier must have a method called Query() that takes no arguments and returns a []PostMetadata and an error
type MetadataQuerier interface {
	Query() ([]PostMetadata, error)
}

// SlugReader defines a contract for any type that can read content based on a slug.
// Anything with a Read(string) (string, error) method can be treated as a SlugReader.
type SlugReader interface {
	Read(slug string) (string, error)
}

// metadata at the top of each blog post
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
