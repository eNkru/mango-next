package web

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed views
var viewsFS embed.FS

//go:embed public/*
var publicFS embed.FS

func Views() fs.FS {
	return viewsFS
}

func Public() fs.FS {
	sub, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("Failed to sub public fs: %v", err)
	}
	return sub
}
