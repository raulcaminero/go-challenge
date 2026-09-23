// Command carrierproxy demonstrates the go-rod backed PolicyProvider by
// logging in to a carrier-style site.
//
// Usage:
//
//	LOGIN_USERNAME=tomsmith LOGIN_PASSWORD='SuperSecretPassword!' go run .
//
// Optional: LOGIN_URL overrides the login page (defaults to DefaultLoginURL),
// CARRIERPROXY_BROWSER_BIN points at a Chrome/Chromium binary, and
// CARRIERPROXY_HEADFUL=1 shows the browser window.
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	if err := run(os.Getenv, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

// run holds main's logic so it can be exercised by tests and so deferred
// cleanup runs on every exit path (log.Fatal would skip it).
func run(getenv func(string) string, out io.Writer) error {
	cfg := DefaultConfig()
	if url := getenv("LOGIN_URL"); url != "" {
		cfg.LoginURL = url
	}
	cfg.BrowserBin = getenv("CARRIERPROXY_BROWSER_BIN")
	cfg.Headful = getenv("CARRIERPROXY_HEADFUL") != ""

	scraper := NewScraper(cfg)
	defer scraper.Close()

	err := scraper.Login(getenv("LOGIN_USERNAME"), getenv("LOGIN_PASSWORD"))
	switch {
	case err == nil:
		fmt.Fprintf(out, "logged in to %s\n", cfg.LoginURL)
	case errors.Is(err, ErrMissingCredentials):
		return errors.New("set LOGIN_USERNAME and LOGIN_PASSWORD")
	case errors.Is(err, ErrInvalidCredentials):
		return errors.New("login rejected: invalid credentials")
	default:
		return fmt.Errorf("login failed: %w", err)
	}

	if _, err := scraper.Policies(); errors.Is(err, ErrNotImplemented) {
		fmt.Fprintln(out, "policy scraping is out of scope for this challenge")
	}
	return nil
}
