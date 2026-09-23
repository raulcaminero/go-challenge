package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Config describes the carrier login page in terms of CSS selectors, so the
// same Scraper drives any site with a conventional username/password form.
type Config struct {
	// LoginURL is the page that hosts the login form.
	LoginURL string

	// Form field selectors.
	UsernameSelector string
	PasswordSelector string
	SubmitSelector   string

	// SuccessSelector matches an element that exists only once logged in
	// (for example a logout link). FailureSelector matches the error banner
	// the site shows when credentials are rejected. Login races the two.
	SuccessSelector string
	FailureSelector string

	// Timeout is the total budget for one Login call: navigation, locating
	// the form, submitting and waiting for the outcome all share it.
	Timeout time.Duration

	// Headful shows the browser window; the zero value runs headless so an
	// unconfigured Config is always safe for servers and CI.
	Headful bool

	// BrowserBin optionally points at a Chrome/Chromium binary. When empty,
	// rod's launcher finds an installed browser or downloads Chromium once
	// into the user cache directory.
	BrowserBin string
}

// DefaultLoginURL is a public site built for automation practice. Its form
// and outcome markers are stable, and its credentials are published on the
// page itself, which makes it a safe default target for this exercise.
const DefaultLoginURL = "https://the-internet.herokuapp.com/login"

// DefaultConfig returns the selectors for DefaultLoginURL.
func DefaultConfig() Config {
	return Config{
		LoginURL:         DefaultLoginURL,
		UsernameSelector: "#username",
		PasswordSelector: "#password",
		SubmitSelector:   `button[type="submit"]`,
		SuccessSelector:  `a[href="/logout"]`,
		FailureSelector:  "#flash.error",
		Timeout:          20 * time.Second,
	}
}

// Scraper is a PolicyProvider backed by a real browser driven with go-rod.
//
// Login is implemented; Policies and DocumentDownload are out of scope for
// this challenge and report ErrNotImplemented once a session exists.
type Scraper struct {
	cfg      Config
	launcher *launcher.Launcher
	browser  *rod.Browser
	loggedIn bool
}

// Compile-time guarantee that Scraper satisfies PolicyProvider.
var _ PolicyProvider = (*Scraper)(nil)

// NewScraper builds a Scraper. No browser is started until Login is called.
func NewScraper(cfg Config) *Scraper {
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultConfig().Timeout
	}
	return &Scraper{cfg: cfg}
}

// Login opens the configured page, submits the credentials and waits for the
// site to reveal the outcome. It returns nil on success, ErrInvalidCredentials
// when the site rejects the credentials, and a sentinel-wrapped error for
// every other failure so callers can branch with errors.Is.
func (s *Scraper) Login(username, password string) error {
	s.loggedIn = false
	if username == "" || password == "" {
		return ErrMissingCredentials
	}
	if err := s.startBrowser(); err != nil {
		return err
	}

	// Open a blank tab and navigate explicitly: unlike creating the target
	// with a URL, Navigate surfaces network failures (connection refused,
	// DNS, TLS) as errors instead of silently landing on Chrome's error page.
	page, err := s.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return fmt.Errorf("carrierproxy: open tab: %w", err)
	}
	page = page.Timeout(s.cfg.Timeout)
	defer page.Close()

	if err := page.Navigate(s.cfg.LoginURL); err != nil {
		return fmt.Errorf("%w: %s (%v)", ErrSiteUnreachable, s.cfg.LoginURL, err)
	}
	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("carrierproxy: load %s: %w", s.cfg.LoginURL, err)
	}

	if err := s.fillForm(page, username, password); err != nil {
		return err
	}

	return s.awaitOutcome(page)
}

// fillForm locates the three form controls, types the credentials and submits.
// A missing control is reported as ErrLoginFormNotFound.
func (s *Scraper) fillForm(page *rod.Page, username, password string) error {
	userField, err := page.Element(s.cfg.UsernameSelector)
	if err != nil {
		return formNotFound(s.cfg.UsernameSelector, err)
	}
	passField, err := page.Element(s.cfg.PasswordSelector)
	if err != nil {
		return formNotFound(s.cfg.PasswordSelector, err)
	}
	submit, err := page.Element(s.cfg.SubmitSelector)
	if err != nil {
		return formNotFound(s.cfg.SubmitSelector, err)
	}

	if err := userField.Input(username); err != nil {
		return fmt.Errorf("carrierproxy: type username: %w", err)
	}
	if err := passField.Input(password); err != nil {
		return fmt.Errorf("carrierproxy: type password: %w", err)
	}
	if err := submit.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("carrierproxy: submit form: %w", err)
	}
	return nil
}

// awaitOutcome races the success and failure markers so a rejected login is
// reported as soon as the site shows its error, without waiting for the full
// timeout.
func (s *Scraper) awaitOutcome(page *rod.Page) error {
	var outcome error
	onSuccess := func(*rod.Element) error {
		s.loggedIn = true
		return nil
	}
	onFailure := func(*rod.Element) error {
		outcome = ErrInvalidCredentials
		return nil
	}

	race := page.Race()
	race.Element(s.cfg.SuccessSelector).Handle(onSuccess)
	race.Element(s.cfg.FailureSelector).Handle(onFailure)
	_, err := race.Do()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%w after %s", ErrLoginOutcomeUnknown, s.cfg.Timeout)
		}
		return fmt.Errorf("%w: %v", ErrLoginOutcomeUnknown, err)
	}
	return outcome
}

// Policies would scrape the policy list. Out of scope for this challenge.
func (s *Scraper) Policies() ([]Policy, error) {
	if !s.loggedIn {
		return nil, ErrNotLoggedIn
	}
	return nil, fmt.Errorf("policies: %w", ErrNotImplemented)
}

// DocumentDownload would stream a policy document. Out of scope for this challenge.
func (s *Scraper) DocumentDownload(downloadKey string) (io.ReadCloser, error) {
	if !s.loggedIn {
		return nil, ErrNotLoggedIn
	}
	return nil, fmt.Errorf("document download %q: %w", downloadKey, ErrNotImplemented)
}

// LoggedIn reports whether the last Login succeeded.
func (s *Scraper) LoggedIn() bool { return s.loggedIn }

// Close shuts the browser down and releases its temporary profile. It is safe
// to call when no browser was started.
func (s *Scraper) Close() error {
	var err error
	if s.browser != nil {
		err = s.browser.Close()
		s.browser = nil
	}
	if s.launcher != nil {
		s.launcher.Kill()
		s.launcher.Cleanup()
		s.launcher = nil
	}
	s.loggedIn = false
	return err
}

// startBrowser launches Chromium and connects rod to it, once per Scraper.
func (s *Scraper) startBrowser() error {
	if s.browser != nil {
		return nil
	}
	l := launcher.New().Headless(!s.cfg.Headful).Leakless(false)
	if s.cfg.BrowserBin != "" {
		l = l.Bin(s.cfg.BrowserBin)
	}
	controlURL, err := l.Launch()
	if err != nil {
		return fmt.Errorf("carrierproxy: launch browser: %w", err)
	}
	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		l.Kill()
		l.Cleanup()
		return fmt.Errorf("carrierproxy: connect browser: %w", err)
	}
	s.launcher = l
	s.browser = browser
	return nil
}

func formNotFound(selector string, cause error) error {
	return fmt.Errorf("%w: %q (%v)", ErrLoginFormNotFound, selector, cause)
}
