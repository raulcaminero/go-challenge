package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoginSuccess(t *testing.T) {
	site := newFixtureSite(t)
	s := newTestScraper(t, site.URL+"/login")
	username, password := testCredentials()

	if err := s.Login(username, password); err != nil {
		t.Fatalf("login: %v", err)
	}
	if !s.LoggedIn() {
		t.Fatal("expected LoggedIn() to be true after a successful login")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	site := newFixtureSite(t)
	s := newTestScraper(t, site.URL+"/login")
	username, _ := testCredentials()

	err := s.Login(username, "definitely-not-the-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("got %v, want ErrInvalidCredentials", err)
	}
	if s.LoggedIn() {
		t.Fatal("LoggedIn() must be false after a rejected login")
	}
}

func TestLoginMissingCredentials(t *testing.T) {
	// No fixture and no browser: validation happens before any browser work.
	s := NewScraper(DefaultConfig())
	t.Cleanup(func() { _ = s.Close() })

	for name, creds := range map[string][2]string{
		"empty username": {"", "secret"},
		"empty password": {"user", ""},
		"both empty":     {"", ""},
	} {
		t.Run(name, func(t *testing.T) {
			if err := s.Login(creds[0], creds[1]); !errors.Is(err, ErrMissingCredentials) {
				t.Fatalf("got %v, want ErrMissingCredentials", err)
			}
		})
	}
}

func TestLoginFormNotFound(t *testing.T) {
	site := newFixtureSite(t)
	s := newTestScraper(t, site.URL+"/noform")
	s.cfg.Timeout = 3 * time.Second // the form never appears; fail fast
	username, password := testCredentials()

	err := s.Login(username, password)
	if !errors.Is(err, ErrLoginFormNotFound) {
		t.Fatalf("got %v, want ErrLoginFormNotFound", err)
	}
}

func TestLoginOutcomeUnknown(t *testing.T) {
	// Point the success/failure markers at selectors the fixture never renders
	// so the outcome race times out.
	site := newFixtureSite(t)
	s := newTestScraper(t, site.URL+"/login")
	s.cfg.SuccessSelector = "#never-rendered-success"
	s.cfg.FailureSelector = "#never-rendered-failure"
	s.cfg.Timeout = 3 * time.Second
	username, password := testCredentials()

	err := s.Login(username, password)
	if !errors.Is(err, ErrLoginOutcomeUnknown) {
		t.Fatalf("got %v, want ErrLoginOutcomeUnknown", err)
	}
}

func TestLoginUnreachableSite(t *testing.T) {
	// A closed port: the browser cannot load the page at all. This must be
	// reported as the site being down, not as a missing form or bad password.
	site := newFixtureSite(t)
	url := site.URL + "/login"
	site.Close()

	s := newTestScraper(t, url)
	username, password := testCredentials()

	err := s.Login(username, password)
	if !errors.Is(err, ErrSiteUnreachable) {
		t.Fatalf("got %v, want ErrSiteUnreachable", err)
	}
	if s.LoggedIn() {
		t.Fatal("LoggedIn() must be false when the site is unreachable")
	}
}

// TestLoginResetsPreviousSession proves a later failed login does not leave a
// stale LoggedIn() == true from an earlier success.
func TestLoginResetsPreviousSession(t *testing.T) {
	site := newFixtureSite(t)
	s := newTestScraper(t, site.URL+"/login")
	username, password := testCredentials()

	if err := s.Login(username, password); err != nil {
		t.Fatalf("first login: %v", err)
	}
	if err := s.Login(username, "wrong-"+password); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("second login: got %v, want ErrInvalidCredentials", err)
	}
	if s.LoggedIn() {
		t.Fatal("LoggedIn() must be reset by a failed login")
	}
}

func TestRunMissingCredentials(t *testing.T) {
	// run() validates before touching a browser, so this stays fast.
	env := map[string]string{}
	err := run(func(k string) string { return env[k] }, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "LOGIN_USERNAME") {
		t.Fatalf("expected a missing-credentials error, got %v", err)
	}
}

func TestRunSuccess(t *testing.T) {
	requireBrowser(t)
	site := newFixtureSite(t)
	username, password := testCredentials()
	env := map[string]string{
		"LOGIN_URL":                site.URL + "/login",
		"LOGIN_USERNAME":           username,
		"LOGIN_PASSWORD":           password,
		"CARRIERPROXY_BROWSER_BIN": os.Getenv(envBrowser),
	}
	var out strings.Builder
	if err := run(func(k string) string { return env[k] }, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(out.String(), "logged in to") || !strings.Contains(out.String(), "out of scope") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunInvalidCredentials(t *testing.T) {
	requireBrowser(t)
	site := newFixtureSite(t)
	username, _ := testCredentials()
	env := map[string]string{
		"LOGIN_URL":                site.URL + "/login",
		"LOGIN_USERNAME":           username,
		"LOGIN_PASSWORD":           "nope",
		"CARRIERPROXY_BROWSER_BIN": os.Getenv(envBrowser),
	}
	err := run(func(k string) string { return env[k] }, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "invalid credentials") {
		t.Fatalf("expected invalid-credentials error, got %v", err)
	}
}

func TestOperationsRequireLogin(t *testing.T) {
	s := NewScraper(DefaultConfig())
	if _, err := s.Policies(); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("Policies before login: got %v, want ErrNotLoggedIn", err)
	}
	if _, err := s.DocumentDownload("doc-1"); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("DocumentDownload before login: got %v, want ErrNotLoggedIn", err)
	}
}

func TestOperationsOutOfScopeAfterLogin(t *testing.T) {
	site := newFixtureSite(t)
	s := newTestScraper(t, site.URL+"/login")
	username, password := testCredentials()
	if err := s.Login(username, password); err != nil {
		t.Fatalf("login: %v", err)
	}
	if _, err := s.Policies(); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("Policies: got %v, want ErrNotImplemented", err)
	}
	if _, err := s.DocumentDownload("doc-1"); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("DocumentDownload: got %v, want ErrNotImplemented", err)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	s := NewScraper(DefaultConfig())
	for i := 0; i < 2; i++ {
		if err := s.Close(); err != nil {
			t.Fatalf("close %d: %v", i, err)
		}
	}
}

func TestNewScraperDefaultsTimeout(t *testing.T) {
	s := NewScraper(Config{LoginURL: "http://example.invalid"})
	if s.cfg.Timeout != DefaultConfig().Timeout {
		t.Fatalf("timeout = %s, want %s", s.cfg.Timeout, DefaultConfig().Timeout)
	}
}

// TestLoginLiveSite runs the real flow against DefaultLoginURL (or LOGIN_URL).
// It needs network access, so it only runs when CARRIERPROXY_LIVE=1.
func TestLoginLiveSite(t *testing.T) {
	if os.Getenv(envLive) == "" {
		t.Skipf("set %s=1 to run against the live site", envLive)
	}
	url := os.Getenv(envLiveURL)
	if url == "" {
		url = DefaultLoginURL
	}
	s := newTestScraper(t, url)
	username, password := testCredentials()

	if err := s.Login(username, password); err != nil {
		t.Fatalf("live login: %v", err)
	}
	if err := newTestScraper(t, url).Login(username, "wrong-"+password); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("live login with a bad password: got %v, want ErrInvalidCredentials", err)
	}
}
