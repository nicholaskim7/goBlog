package handler

import (
	"html/template"
	"net/http"

	"github.com/nicholaskim7/go-blog/internal/post"
)

func RegisterRoutes(postReader post.FileReader) *http.ServeMux {
	// Create a new request router (also called a "serve mux").
	mux := http.NewServeMux()

	// Parse all templates
	postTemplate := template.Must(template.ParseFiles("tmpl/post.gohtml"))
	indexTemplate := template.Must(template.ParseFiles("tmpl/index.gohtml"))
	writeTemplate := template.Must(template.ParseFiles("tmpl/write.gohtml"))

	// register handlers
	mux.HandleFunc("GET /", IndexHandler(postReader, indexTemplate))
	mux.HandleFunc("GET /posts/{slug}", PostHandler(postReader, postTemplate))
	mux.HandleFunc("GET /write", WriteHandler(writeTemplate))
	mux.HandleFunc("POST /submit", SubmitHandler())

	// handler for static files
	mux.Handle("GET /public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	return mux
}
