package filestore_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/gloveboxhq/glovebox-go-code-challenge/filestore"
	"github.com/gloveboxhq/glovebox-go-code-challenge/filestore/mock"
	"github.com/gloveboxhq/glovebox-go-code-challenge/filestore/s3"
)

func newMock() *mock.Client {
	return mock.NewClient(mock.Config{Bucket: mock.Bucket{Name: "test"}, BasePath: "tenant-1"})
}

// mustSet seeds an object and fails the test on error.
func mustSet(t *testing.T, client *mock.Client, name string, content []byte, contentType string) {
	t.Helper()
	if err := client.Set(context.Background(), name, content, contentType); err != nil {
		t.Fatalf("set %q: %v", name, err)
	}
}

// readAll reads an object fully and returns its bytes and content type.
func readAll(t *testing.T, client *mock.Client, name string) ([]byte, string) {
	t.Helper()
	rc, contentType, err := client.Get(context.Background(), name)
	if err != nil {
		t.Fatalf("get %q: %v", name, err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read %q: %v", name, err)
	}
	return data, contentType
}

func TestMockPurge(t *testing.T) {
	client := newMock()
	if err := client.Purge(context.Background(), "missing.txt"); err != nil {
		t.Fatalf("purging a missing file: %v", err)
	}
}

// TestFileProviderContract pins the error contract every provider must keep:
// missing files surface filestore.ErrNotFound through errors.Is.
func TestFileProviderContract(t *testing.T) {
	_, _, err := newMock().Get(context.Background(), "missing.txt")
	if !errors.Is(err, filestore.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMockSetGetRoundTrip(t *testing.T) {
	client := newMock()
	want := []byte("hello, glovebox")
	mustSet(t, client, "docs/policy.pdf", want, "application/pdf")

	// Reading twice must work: handing out a reader never consumes the object.
	for i := 0; i < 2; i++ {
		got, contentType := readAll(t, client, "docs/policy.pdf")
		if !bytes.Equal(got, want) {
			t.Fatalf("read %d: got %q, want %q", i, got, want)
		}
		if contentType != "application/pdf" {
			t.Fatalf("read %d: got content type %q, want %q", i, contentType, "application/pdf")
		}
	}
}

func TestMockSetDefaultsContentType(t *testing.T) {
	client := newMock()
	mustSet(t, client, "blob", []byte{1, 2, 3}, "")
	if _, contentType := readAll(t, client, "blob"); contentType != "application/octet-stream" {
		t.Fatalf("got content type %q, want application/octet-stream", contentType)
	}
}

func TestMockCopy(t *testing.T) {
	const (
		src = "in/original.pdf"
		dst = "out/duplicate.pdf"
	)
	content := []byte("policy document body")

	tests := map[string]struct {
		seed    func(t *testing.T, c *mock.Client)
		src     string
		dst     string
		wantErr error
	}{
		"success": {
			seed: func(t *testing.T, c *mock.Client) { mustSet(t, c, src, content, "application/pdf") },
			src:  src,
			dst:  dst,
		},
		"missing source": {
			seed:    func(*testing.T, *mock.Client) {},
			src:     "in/nope.pdf",
			dst:     dst,
			wantErr: filestore.ErrNotFound,
		},
		"destination already exists": {
			seed: func(t *testing.T, c *mock.Client) {
				mustSet(t, c, src, content, "application/pdf")
				mustSet(t, c, dst, []byte("do not clobber me"), "text/plain")
			},
			src:     src,
			dst:     dst,
			wantErr: filestore.ErrFileExists,
		},
		"copy onto itself is a conflict": {
			seed:    func(t *testing.T, c *mock.Client) { mustSet(t, c, src, content, "application/pdf") },
			src:     src,
			dst:     src,
			wantErr: filestore.ErrFileExists,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := newMock()
			tc.seed(t, client)

			err := client.Copy(context.Background(), tc.src, tc.dst)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("copy: %v", err)
			}

			gotDst, dstType := readAll(t, client, tc.dst)
			if !bytes.Equal(gotDst, content) || dstType != "application/pdf" {
				t.Fatalf("destination = (%q, %q), want (%q, %q)", gotDst, dstType, content, "application/pdf")
			}
			gotSrc, _ := readAll(t, client, tc.src)
			if !bytes.Equal(gotSrc, content) {
				t.Fatalf("source changed after copy: %q", gotSrc)
			}
		})
	}
}

// TestMockCopyIsIndependent guards against the aliasing bug where source and
// destination share one buffer: purging the source must not touch the copy.
func TestMockCopyIsIndependent(t *testing.T) {
	client := newMock()
	mustSet(t, client, "a.txt", []byte("shared?"), "text/plain")
	if err := client.Copy(context.Background(), "a.txt", "b.txt"); err != nil {
		t.Fatalf("copy: %v", err)
	}
	if err := client.Purge(context.Background(), "a.txt"); err != nil {
		t.Fatalf("purge: %v", err)
	}
	if _, _, err := client.Get(context.Background(), "a.txt"); !errors.Is(err, filestore.ErrNotFound) {
		t.Fatalf("source should be gone, got %v", err)
	}
	if got, _ := readAll(t, client, "b.txt"); string(got) != "shared?" {
		t.Fatalf("copy lost its content: %q", got)
	}
}

// TestMockDestinationConflictContract ensures the conflict rule is shared by
// every operation that writes to a new key, not just Copy.
func TestMockDestinationConflictContract(t *testing.T) {
	client := newMock()
	mustSet(t, client, "src", []byte("x"), "")
	mustSet(t, client, "taken", []byte("y"), "")

	if err := client.Move(context.Background(), "src", "taken"); !errors.Is(err, filestore.ErrFileExists) {
		t.Fatalf("move onto existing key: got %v, want ErrFileExists", err)
	}
	if err := client.Move(context.Background(), "ghost", "free"); !errors.Is(err, filestore.ErrNotFound) {
		t.Fatalf("move missing key: got %v, want ErrNotFound", err)
	}
	if got, _ := readAll(t, client, "taken"); string(got) != "y" {
		t.Fatalf("existing destination was overwritten: %q", got)
	}
}

func TestMockHonoursCancelledContext(t *testing.T) {
	client := newMock()
	mustSet(t, client, "a", []byte("a"), "")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.Copy(ctx, "a", "b"); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

// TestS3StubContract keeps the stub honest with the shared error contract so
// swapping providers in tests does not change error handling.
func TestS3StubContract(t *testing.T) {
	client := s3.NewClient(s3.Config{Bucket: "stub"})
	if err := client.Copy(context.Background(), "src", "dst"); !errors.Is(err, filestore.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}
