package filestore

import (
	"context"
	"io"
	"time"
)

// FileProvider is the storage abstraction the application talks to. Callers
// never see provider-specific details; every implementation must honour the
// same error contract:
//
//   - reading, moving or copying a filename that does not exist returns an
//     error that satisfies errors.Is(err, ErrNotFound) (Purge is idempotent
//     and succeeds on a missing file)
//   - moving or copying onto an existing destination returns an error that
//     satisfies errors.Is(err, ErrFileExists); nothing is ever overwritten
type FileProvider interface {
	Get(ctx context.Context, filename string) (io.ReadCloser, string, error)
	Set(ctx context.Context, filename string, fileBytes []byte, contentType string) error
	Purge(ctx context.Context, filename string) error
	Move(ctx context.Context, oldFilename, newFilename string) error

	// Copy duplicates the object stored at srcFilename into dstFilename,
	// preserving its content and content type. The source is left untouched.
	//
	// It returns ErrNotFound when the source does not exist and ErrFileExists
	// when the destination is already taken; Copy never overwrites.
	Copy(ctx context.Context, srcFilename, dstFilename string) error
}

type PresignedFileProvider interface {
	FileProvider
	GetPresignedURL(ctx context.Context, filename string, expireAfter time.Duration) (string, error)
}
