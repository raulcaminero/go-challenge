package mock

import (
	"bytes"
	"errors"
)

// MemoryFile is an in-memory object: immutable content plus its content type.
//
// Reading is done through a private cursor so the stored bytes are never
// consumed or mutated; Close only invalidates this handle. Callers obtain an
// independent handle per Get, so closing one never affects the stored object
// or other readers.
type MemoryFile struct {
	data        []byte
	contentType string
	cursor      *bytes.Reader
	closed      bool
}

// NewMemoryFile stores a private copy of data so later changes to the caller's
// slice cannot leak into the bucket.
func NewMemoryFile(data []byte, contentType string) *MemoryFile {
	return &MemoryFile{data: bytes.Clone(data), contentType: contentType}
}

func (f *MemoryFile) Read(p []byte) (int, error) {
	if f.closed {
		return 0, errors.New("file is closed")
	}
	if f.cursor == nil {
		f.cursor = bytes.NewReader(f.data)
	}
	return f.cursor.Read(p)
}

func (f *MemoryFile) Close() error {
	f.closed = true
	f.cursor = nil
	return nil
}

// Bytes returns a copy of the stored content.
func (f *MemoryFile) Bytes() []byte { return bytes.Clone(f.data) }

// ContentType returns the MIME type recorded when the object was stored.
func (f *MemoryFile) ContentType() string { return f.contentType }

// clone returns a fresh, unread, unclosed handle over an independent copy of
// the content. It is what Get hands out and what Copy stores.
func (f *MemoryFile) clone() *MemoryFile { return NewMemoryFile(f.data, f.contentType) }
