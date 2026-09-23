// Package s3 is the S3-shaped filestore.FileProvider. In this extraction the
// AWS SDK is not vendored, so the client is a stub that documents the exact
// calls and error translation a production implementation performs.
package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gloveboxhq/glovebox-go-code-challenge/filestore"
)

type Config struct {
	Bucket string
}

type Client struct {
	bucket string
}

// Compile-time guarantee that the S3 client satisfies both provider interfaces.
var _ filestore.PresignedFileProvider = (*Client)(nil)

func NewClient(config Config) *Client { return &Client{bucket: config.Bucket} }

func (c *Client) Get(_ context.Context, _ string) (io.ReadCloser, string, error) {
	return nil, "", filestore.ErrNotFound
}

func (c *Client) Set(_ context.Context, _ string, _ []byte, _ string) error { return nil }

func (c *Client) Purge(_ context.Context, _ string) error { return nil }

func (c *Client) Move(_ context.Context, _, _ string) error { return nil }

// Copy duplicates srcFilename into dstFilename server-side.
//
// Against real S3 this is:
//  1. HeadObject(dst) - a 200 means the key is taken -> filestore.ErrFileExists
//     (S3 CopyObject overwrites by default, so the check is ours to make).
//  2. CopyObject(CopySource: bucket/src, Key: dst) with
//     MetadataDirective=COPY so the content type travels with the object.
//  3. A NoSuchKey / 404 from CopyObject -> filestore.ErrNotFound.
//
// The stub holds no objects, so - consistently with Get - every source is
// reported missing.
func (c *Client) Copy(_ context.Context, srcFilename, _ string) error {
	return fmt.Errorf("copy %q: %w", srcFilename, filestore.ErrNotFound)
}

func (c *Client) GetPresignedURL(_ context.Context, filename string, _ time.Duration) (string, error) {
	return "https://storage.example/" + c.bucket + "/" + filename, nil
}
