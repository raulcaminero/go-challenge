package main

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/gloveboxhq/glovebox-go-code-challenge/filestore"
	"github.com/gloveboxhq/glovebox-go-code-challenge/filestore/mock"
)

func main() {
	ctx := context.Background()
	client := mock.NewClient(mock.Config{Bucket: mock.Bucket{Name: "local"}})

	if err := client.Set(ctx, "example.txt", []byte("hello"), "text/plain"); err != nil {
		log.Fatal(err)
	}
	if err := client.Copy(ctx, "example.txt", "example-copy.txt"); err != nil {
		log.Fatal(err)
	}

	// A second copy onto the same key is refused rather than overwritten.
	if err := client.Copy(ctx, "example.txt", "example-copy.txt"); !errors.Is(err, filestore.ErrFileExists) {
		log.Fatalf("expected ErrFileExists, got %v", err)
	}

	rc, contentType, err := client.Get(ctx, "example-copy.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()
	body, err := io.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("copied %q (%s): %s", "example-copy.txt", contentType, body)
	log.Print("filestore challenge ready")
}
