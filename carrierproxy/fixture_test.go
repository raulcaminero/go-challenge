package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Credentials come from the environment so the suite can be pointed at any
// account. The defaults are the published credentials of DefaultLoginURL,
// so one pair of variables drives both the offline fixture and the live site.
const (
	envUsername = "LOGIN_USERNAME"
	envPassword = "LOGIN_PASSWORD"
	envLiveURL  = "LOGIN_URL"
	envLive     = "CARRIERPROXY_LIVE"
	envBrowser  = "CARRIERPROXY_BROWSER_BIN"

	defaultUsername = "tomsmith"
	defaultPassword = "SuperSecretPassword!"
)

func testCredentials() (username, password string) {
	username, password = os.Getenv(envUsername), os.Getenv(envPassword)
	if username == "" {
		username = defaultUsername
	}
	if password == "" {
		password = defaultPassword
	}
	return username, password
}

// newFixtureSite serves a minimal carrier-style login flow with the same
// markers as DefaultLoginURL, so the suite runs offline and deterministically:
//
//	GET  /login   -> form (#username, #password, button[type=submit])
//	POST /login   -> 303 to /secure on valid credentials, else the form with
//	                 the #flash.error banner
//	GET  /secure  -> page containing a[href="/logout"]
//	GET  /noform  -> a page with no login form at all
func newFixtureSite(t *testing.T) *httptest.Server {
	t.Helper()
	validUser, validPass := testCredentials()

	const form = `<!doctype html><html><body>
<h2>Login Page</h2>%s
<form id="login" method="post" action="/login">
  <label for="username">Username</label><input type="text" id="username" name="username">
  <label for="password">Password</label><input type="password" id="password" name="password">
  <button type="submit">Login</button>
</form></body></html>`

	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(w, form, "")
		case http.MethodPost:
			if r.FormValue("username") == validUser && r.FormValue("password") == validPass {
				http.Redirect(w, r, "/secure", http.StatusSeeOther)
				return
			}
			fmt.Fprintf(w, form, `<div id="flash" class="flash error">Your username is invalid!</div>`)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/secure", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<!doctype html><html><body>
<div id="flash" class="flash success">You logged into a secure area!</div>
<a class="button" href="/logout">Logout</a></body></html>`)
	})
	mux.HandleFunc("/noform", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<!doctype html><html><body><h2>Maintenance</h2><p>Back soon.</p></body></html>`)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// requireBrowser skips browser-backed tests under -short. Without a browser on
// the machine rod downloads Chromium on first use, which is unwanted in quick
// or offline runs; the pure-logic tests still execute.
func requireBrowser(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("browser tests skipped with -short")
	}
}

// newTestScraper points a Scraper at the fixture and guarantees the browser is
// shut down when the test ends.
func newTestScraper(t *testing.T, loginURL string) *Scraper {
	t.Helper()
	requireBrowser(t)
	cfg := DefaultConfig()
	cfg.LoginURL = loginURL
	cfg.BrowserBin = os.Getenv(envBrowser)
	s := NewScraper(cfg)
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Logf("close scraper: %v", err)
		}
	})
	return s
}
