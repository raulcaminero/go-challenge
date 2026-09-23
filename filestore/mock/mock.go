// Package mock is an in-memory filestore.FileProvider used by tests and the
// local demo. It mirrors the semantics a real object store exposes: keys are
// slash-separated paths, objects are immutable blobs with a content type, and
// the sentinel errors from the filestore package are always preserved so
// callers can rely on errors.Is.
package mock

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/gloveboxhq/glovebox-go-code-challenge/filestore"
)

const defaultContentType = "application/octet-stream"

type Bucket struct {
	Name    string
	Objects map[string]*MemoryFile
}

type Config struct {
	Bucket   Bucket
	BasePath string
}

type Client struct {
	bucket   Bucket
	basePath string
}

// Compile-time guarantee that the mock satisfies both provider interfaces.
var _ filestore.PresignedFileProvider = (*Client)(nil)

func NewClient(config Config) *Client {
	if config.Bucket.Objects == nil {
		config.Bucket.Objects = map[string]*MemoryFile{}
	}
	return &Client{bucket: config.Bucket, basePath: config.BasePath}
}

// key builds the object key. Object-store keys are always '/'-separated, so
// path.Join is used rather than filepath.Join (which is OS-specific).
func (c *Client) key(filename string) string { return path.Join(c.basePath, filename) }

// lookup fetches an object or returns a wrapped filestore.ErrNotFound that
// names the operation and key for easier debugging.
func (c *Client) lookup(op, filename string) (*MemoryFile, error) {
	file, ok := c.bucket.Objects[c.key(filename)]
	if !ok {
		return nil, fmt.Errorf("%s %q: %w", op, filename, filestore.ErrNotFound)
	}
	return file, nil
}

// requireAbsent enforces the destination-conflict contract shared by Move and Copy.
func (c *Client) requireAbsent(op, filename string) error {
	if _, ok := c.bucket.Objects[c.key(filename)]; ok {
		return fmt.Errorf("%s %q: %w", op, filename, filestore.ErrFileExists)
	}
	return nil
}

func (c *Client) Get(ctx context.Context, filename string) (io.ReadCloser, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	file, err := c.lookup("get", filename)
	if err != nil {
		return nil, "", err
	}
	// Hand out an independent handle so the caller can read and Close it
	// without consuming or invalidating the stored object.
	return file.clone(), file.ContentType(), nil
}

func (c *Client) Set(ctx context.Context, filename string, fileBytes []byte, contentType string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if contentType == "" {
		contentType = defaultContentType
	}
	c.bucket.Objects[c.key(filename)] = NewMemoryFile(fileBytes, contentType)
	return nil
}

// Purge is idempotent: deleting a missing object is not an error, matching
// S3 DeleteObject semantics.
func (c *Client) Purge(ctx context.Context, filename string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(c.bucket.Objects, c.key(filename))
	return nil
}

func (c *Client) Move(ctx context.Context, oldFilename, newFilename string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := c.lookup("move", oldFilename)
	if err != nil {
		return err
	}
	if err := c.requireAbsent("move", newFilename); err != nil {
		return err
	}
	c.bucket.Objects[c.key(newFilename)] = file
	delete(c.bucket.Objects, c.key(oldFilename))
	return nil
}

// Copy stores an independent duplicate of the source object under the new
// key. Copying onto an existing key (including onto itself) is refused with
// filestore.ErrFileExists so callers can never silently clobber data.
func (c *Client) Copy(ctx context.Context, srcFilename, dstFilename string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := c.lookup("copy", srcFilename)
	if err != nil {
		return err
	}
	if err := c.requireAbsent("copy", dstFilename); err != nil {
		return err
	}
	c.bucket.Objects[c.key(dstFilename)] = file.clone()
	return nil
}

func (c *Client) GetPresignedURL(ctx context.Context, filename string, _ time.Duration) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if _, err := c.lookup("presign", filename); err != nil {
		return "", err
	}
	return "http://not-a-real-presigned-url.example/" + c.key(filename), nil
}
