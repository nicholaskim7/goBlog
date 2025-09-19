package post

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
)

type FileReader struct {
	// Directory to find blog posts in
	Dir string
}

// Read finds and reads a markdown file using the slug.
// this method has the same signature as the Read method in the SlugReader interface
func (fr FileReader) Read(slug string) (string, error) {
	// include directory to path
	slugPath := filepath.Join(fr.Dir, slug+".md")
	// open markdown file
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

// Query finds metadata for all posts.
// this method has the same signature as the Query method in the MetadataQuerier interface
// you can treat a FileReader object as if it were a MetadataQuerier
// in other words the FileReader struct implements the query function and metaDataQuerier interface uses it
func (fr FileReader) Query() ([]PostMetadata, error) {
	// include directory to path
	postsPath := filepath.Join(fr.Dir, "*.md")
	// Glob all the markdown files in the current directory
	filenames, err := filepath.Glob(postsPath)
	if err != nil {
		return nil, fmt.Errorf("querying for files: %w", err)
	}
	// slice to hold all post metadata
	var posts []PostMetadata
	for _, filename := range filenames {
		// Open each file
		f, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("opening file %q: %w", filename, err)
		}
		defer f.Close()
		var post PostMetadata
		//parse frontmatter metadata into post
		_, err = frontmatter.Parse(f, &post)
		if err != nil {
			return nil, fmt.Errorf("parsing frontmatter for file %s: %w", filename, err)
		}
		// get the base filename and remove the .md suffix to get the slug
		post.Slug = strings.TrimSuffix(filepath.Base(filename), ".md")
		// append metadata of post to posts slice
		posts = append(posts, post)
	}
	return posts, nil
}
