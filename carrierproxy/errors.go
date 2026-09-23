package main

import "errors"

// Sentinel errors let callers branch on the outcome of a scrape with
// errors.Is instead of parsing messages.
var (
	// ErrMissingCredentials is returned before any browser work when the
	// username or password is empty.
	ErrMissingCredentials = errors.New("carrierproxy: username and password are required")

	// ErrInvalidCredentials means the carrier site rejected the login.
	ErrInvalidCredentials = errors.New("carrierproxy: invalid credentials")

	// ErrSiteUnreachable means the browser could not load the login page at
	// all (DNS failure, connection refused, TLS error, ...).
	ErrSiteUnreachable = errors.New("carrierproxy: login page unreachable")

	// ErrLoginFormNotFound means the page loaded but the expected login form
	// fields never appeared - typically a changed carrier UI or a wrong URL.
	ErrLoginFormNotFound = errors.New("carrierproxy: login form not found")

	// ErrLoginOutcomeUnknown means the form was submitted but neither the
	// success nor the failure marker appeared before the timeout.
	ErrLoginOutcomeUnknown = errors.New("carrierproxy: could not determine login outcome")

	// ErrNotLoggedIn is returned by operations that require a session.
	ErrNotLoggedIn = errors.New("carrierproxy: not logged in")

	// ErrNotImplemented marks PolicyProvider operations outside this
	// challenge's scope.
	ErrNotImplemented = errors.New("carrierproxy: not implemented")
)
