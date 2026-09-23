# Carrierproxy Scraper Challenge

This challenge is to create a partial implementation of the `PolicyProvider` interface. Normally this interface would be used to scrape policies from a carrier website but in this case any website with a login form can be used as the target.

**Instructions:**

* [ ] Implement the `PolicyProvider`'s `Login` method against a website of your choice using the [go-rod](https://github.com/go-rod/rod) scraping library.
* [ ] Create test(s) with full coverage of your code that accept environment variables for the credentials (login & password).
* [ ] Document your code and usage instructions.

## Implementation

`Scraper` (`scraper.go`) implements `PolicyProvider` with a real Chromium
instance driven by go-rod. It is configured with CSS selectors (`Config`), so
the same code logs in to any site with a conventional username/password form.

`Login(username, password)`:

1. validates the credentials are non-empty (`ErrMissingCredentials`) before
   starting a browser;
2. launches headless Chromium once per `Scraper`, opens a tab and navigates to
   `Config.LoginURL` (`ErrSiteUnreachable` if the page cannot be loaded);
3. locates the username, password and submit controls (`ErrLoginFormNotFound`
   if any is missing), types the credentials and submits;
4. races the success marker (an element only present when logged in, e.g. a
   logout link) against the failure marker (the site's error banner) and
   returns `nil` or `ErrInvalidCredentials` accordingly. If neither appears
   it returns `ErrLoginOutcomeUnknown`.

`Config.Timeout` (default 20s) is the total budget for one `Login` call.

Every failure is wrapped around a sentinel in `errors.go` so callers use
`errors.Is` rather than string matching. `Policies` and `DocumentDownload`
require a session (`ErrNotLoggedIn`) and otherwise return `ErrNotImplemented`:
scraping them is outside the scope of this challenge. `Close` shuts the browser
down and removes its temporary profile.

### Target site

`DefaultConfig()` targets <https://the-internet.herokuapp.com/login>, a public
site built for automation practice. Its credentials are published on the page
(`tomsmith` / `SuperSecretPassword!`), its markup is stable, and automating it
breaks nobody's terms of service - which is why it was chosen over a real
carrier or consumer site.

## Requirements

* Go 1.22+
* Chrome or Chromium. If none is installed, go-rod's launcher downloads a
  matching Chromium build into the user cache directory on first run (network
  access needed once). To use a specific binary set `CARRIERPROXY_BROWSER_BIN`.

## Usage

```bash
cd carrierproxy

# log in to the default site
LOGIN_USERNAME=tomsmith LOGIN_PASSWORD='SuperSecretPassword!' go run .

# another site with the same form conventions
LOGIN_URL=https://example.com/login LOGIN_USERNAME=... LOGIN_PASSWORD=... go run .

# watch the browser instead of running headless
CARRIERPROXY_HEADFUL=1 LOGIN_USERNAME=tomsmith LOGIN_PASSWORD='SuperSecretPassword!' go run .
```

Programmatic use:

```go
cfg := DefaultConfig()            // or fill Config with your own selectors
scraper := NewScraper(cfg)
defer scraper.Close()

if err := scraper.Login(user, pass); errors.Is(err, ErrInvalidCredentials) {
    // wrong username/password
}
```

## Tests

The suite is offline and deterministic by default: `fixture_test.go` starts an
`httptest` server that reproduces the login flow and markers of the default
site, and every `Login` path is exercised against it - success, rejected
credentials, missing credentials, missing form, unknown outcome, unreachable
site, session reset - plus the session guards on `Policies`/`DocumentDownload`
and the `run` entry point behind `main`. Browser-backed tests are skipped with
`go test -short`, leaving the pure-logic tests.

Credentials are read from the environment (defaults are the public ones):

| Variable | Purpose | Default |
|---|---|---|
| `LOGIN_USERNAME` | username accepted by the fixture and used for the live site | `tomsmith` |
| `LOGIN_PASSWORD` | password accepted by the fixture and used for the live site | `SuperSecretPassword!` |
| `LOGIN_URL` | live login page (live test and `go run`) | `https://the-internet.herokuapp.com/login` |
| `CARRIERPROXY_LIVE` | set to `1` to also run the live-site test | unset (skipped) |
| `CARRIERPROXY_BROWSER_BIN` | path to a Chrome/Chromium binary | auto-detect / download |

```bash
go test ./... -cover

# quick, browser-free subset
go test ./... -short

# include the real site
CARRIERPROXY_LIVE=1 LOGIN_USERNAME=tomsmith LOGIN_PASSWORD='SuperSecretPassword!' go test ./... -run Live -v
```
