package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gloveboxhq/glovebox-go-code-challenge/comms/email"
	"github.com/gloveboxhq/glovebox-go-code-challenge/comms/email/mockemail"
	"github.com/gloveboxhq/glovebox-go-code-challenge/comms/handlers"
)

func TestAddPolicyVehicle(t *testing.T) {

	t.Parallel()

	type testCase struct {
		method       string
		payload      handlers.AddPolicyVehicleReq
		expectTplID  email.TplID
		expectStatus int
	}

	testCases := map[string]testCase{
		"pass": {
			method: http.MethodPost,
			payload: handlers.AddPolicyVehicleReq{
				EmailTo: "foo@bar.com",
				Message: json.RawMessage(`{"foo":"bar"}`),
			},
			expectTplID:  email.TplAddPolicyVehicle,
			expectStatus: http.StatusOK,
		},
		"fail invalid method": {
			method:       http.MethodGet,
			payload:      handlers.AddPolicyVehicleReq{},
			expectStatus: http.StatusMethodNotAllowed,
		},
	}

	testFactory := func(tc testCase) func(*testing.T) {
		return func(t *testing.T) {

			payload, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("could nor marshal payload to json: %v", err)
			}

			testEmail := mockemail.NewClient()
			defer testEmail.FlushSendLogs()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/api/comms/add-policy-vehicle", bytes.NewReader(payload))

			handlers.AddPolicyVehicle(testEmail)(w, req)

			resp := w.Result()

			if resp.StatusCode != tc.expectStatus {
				t.Fatalf("expected status %v but got %v", tc.expectStatus, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusOK {

				if testEmail.SendLogs().IsEmpty() {
					t.Fatalf("expected email log but got empty")
				}

				lastEmail := testEmail.SendLogs().Last()

				if lastEmail.ExtractTo() != tc.payload.EmailTo {
					t.Fatalf("expected to %v but got %v", tc.payload.EmailTo, lastEmail.ExtractTo())
				}

				if string(lastEmail.ExtractMessage()) != string(tc.payload.Message) {
					t.Fatalf("expected message %v but got %v", tc.payload.Message, lastEmail.ExtractMessage())
				}

				if lastEmail.ExtractTplID() != tc.expectTplID {
					t.Fatalf("expected tpl %v but got %v", tc.expectTplID, lastEmail.ExtractTplID())
				}
			}
		}
	}

	for name, tc := range testCases {
		t.Run(name, testFactory(tc))
	}
}

func TestAddPolicyDriver(t *testing.T) {

	t.Parallel()

	type testCase struct {
		method       string
		payload      handlers.AddPolicyDriverReq
		expectTplID  email.TplID
		expectStatus int
	}

	testCases := map[string]testCase{
		"pass": {
			method: http.MethodPost,
			payload: handlers.AddPolicyDriverReq{
				EmailTo: "foo@bar.com",
				Message: json.RawMessage(`{"foo":"bar"}`),
			},
			expectTplID:  email.TplAddPolicyDriver,
			expectStatus: http.StatusOK,
		},
		"fail invalid method": {
			method:       http.MethodGet,
			payload:      handlers.AddPolicyDriverReq{},
			expectStatus: http.StatusMethodNotAllowed,
		},
	}

	testFactory := func(tc testCase) func(*testing.T) {
		return func(t *testing.T) {

			payload, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("could nor marshal payload to json: %v", err)
			}

			testEmail := mockemail.NewClient()
			defer testEmail.FlushSendLogs()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/api/comms/add-policy-driver", bytes.NewReader(payload))

			handlers.AddPolicyDriver(testEmail)(w, req)

			resp := w.Result()

			if resp.StatusCode != tc.expectStatus {
				t.Fatalf("expected status %v but got %v", tc.expectStatus, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusOK {

				if testEmail.SendLogs().IsEmpty() {
					t.Fatalf("expected email log but got empty")
				}

				lastEmail := testEmail.SendLogs().Last()

				if lastEmail.ExtractTo() != tc.payload.EmailTo {
					t.Fatalf("expected to %v but got %v", tc.payload.EmailTo, lastEmail.ExtractTo())
				}

				if string(lastEmail.ExtractMessage()) != string(tc.payload.Message) {
					t.Fatalf("expected message %v but got %v", tc.payload.Message, lastEmail.ExtractMessage())
				}

				if lastEmail.ExtractTplID() != tc.expectTplID {
					t.Fatalf("expected tpl %v but got %v", tc.expectTplID, lastEmail.ExtractTplID())
				}
			}
		}
	}

	for name, tc := range testCases {
		t.Run(name, testFactory(tc))
	}
}

func TestAddPolicyAddress(t *testing.T) {

	t.Parallel()

	type testCase struct {
		method       string
		payload      handlers.AddPolicyAddressReq
		expectTplID  email.TplID
		expectStatus int
	}

	testCases := map[string]testCase{
		"pass": {
			method: http.MethodPost,
			payload: handlers.AddPolicyAddressReq{
				EmailTo: "foo@bar.com",
				Message: json.RawMessage(`{"foo":"bar"}`),
			},
			expectTplID:  email.TplAddPolicyAddress,
			expectStatus: http.StatusOK,
		},
		"fail invalid method": {
			method:       http.MethodGet,
			payload:      handlers.AddPolicyAddressReq{},
			expectStatus: http.StatusMethodNotAllowed,
		},
	}

	testFactory := func(tc testCase) func(*testing.T) {
		return func(t *testing.T) {

			payload, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("could nor marshal payload to json: %v", err)
			}

			testEmail := mockemail.NewClient()
			defer testEmail.FlushSendLogs()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/api/comms/add-policy-address", bytes.NewReader(payload))

			handlers.AddPolicyAddress(testEmail)(w, req)

			resp := w.Result()

			if resp.StatusCode != tc.expectStatus {
				t.Fatalf("expected status %v but got %v", tc.expectStatus, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusOK {

				if testEmail.SendLogs().IsEmpty() {
					t.Fatalf("expected email log but got empty")
				}

				lastEmail := testEmail.SendLogs().Last()

				if lastEmail.ExtractTo() != tc.payload.EmailTo {
					t.Fatalf("expected to %v but got %v", tc.payload.EmailTo, lastEmail.ExtractTo())
				}

				if string(lastEmail.ExtractMessage()) != string(tc.payload.Message) {
					t.Fatalf("expected message %v but got %v", tc.payload.Message, lastEmail.ExtractMessage())
				}

				if lastEmail.ExtractTplID() != tc.expectTplID {
					t.Fatalf("expected tpl %v but got %v", tc.expectTplID, lastEmail.ExtractTplID())
				}
			}
		}
	}

	for name, tc := range testCases {
		t.Run(name, testFactory(tc))
	}
}

func TestAddPolicyCoverage(t *testing.T) {

	t.Parallel()

	type testCase struct {
		method       string
		body         string // raw body so malformed JSON can be exercised
		expectStatus int
		expectTo     string
		expectCC     []string
		expectSend   bool
	}

	testCases := map[string]testCase{
		"pass with cc": {
			method:       http.MethodPost,
			body:         `{"email_to":"holder@bar.com","email_cc":["agent@bar.com","spouse@bar.com"],"message":{"coverage":"collision"}}`,
			expectStatus: http.StatusOK,
			expectTo:     "holder@bar.com",
			expectCC:     []string{"agent@bar.com", "spouse@bar.com"},
			expectSend:   true,
		},
		"pass without cc keeps existing contract": {
			method:       http.MethodPost,
			body:         `{"email_to":"holder@bar.com","message":{"coverage":"collision"}}`,
			expectStatus: http.StatusOK,
			expectTo:     "holder@bar.com",
			expectCC:     []string{},
			expectSend:   true,
		},
		"cc is cleaned: blanks, duplicates and the To address are dropped": {
			method:       http.MethodPost,
			body:         `{"email_to":"holder@bar.com","email_cc":["", " agent@bar.com ","AGENT@bar.com","Holder@Bar.com"],"message":{}}`,
			expectStatus: http.StatusOK,
			expectTo:     "holder@bar.com",
			expectCC:     []string{"agent@bar.com"},
			expectSend:   true,
		},
		"fail missing email_to": {
			method:       http.MethodPost,
			body:         `{"email_cc":["agent@bar.com"],"message":{}}`,
			expectStatus: http.StatusBadRequest,
		},
		"fail malformed json": {
			method:       http.MethodPost,
			body:         `{"email_to":`,
			expectStatus: http.StatusBadRequest,
		},
		"fail invalid method": {
			method:       http.MethodGet,
			body:         `{"email_to":"holder@bar.com","message":{}}`,
			expectStatus: http.StatusMethodNotAllowed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			testEmail := mockemail.NewClient()
			defer testEmail.FlushSendLogs()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/api/comms/add-policy-coverage", strings.NewReader(tc.body))

			handlers.AddPolicyCoverage(testEmail)(w, req)

			resp := w.Result()

			if resp.StatusCode != tc.expectStatus {
				t.Fatalf("expected status %v but got %v", tc.expectStatus, resp.StatusCode)
			}

			if !tc.expectSend {
				if !testEmail.SendLogs().IsEmpty() {
					t.Fatalf("expected no email on a %d response but %d were sent", resp.StatusCode, len(testEmail.SendLogs()))
				}
				return
			}

			logs := testEmail.SendLogs()
			if len(logs) != 1 {
				t.Fatalf("expected exactly one email (To + CC travel together), got %d", len(logs))
			}
			lastEmail := logs.Last()

			if lastEmail.ExtractTo() != tc.expectTo {
				t.Fatalf("expected to %v but got %v", tc.expectTo, lastEmail.ExtractTo())
			}

			if got := lastEmail.ExtractCC(); !reflect.DeepEqual(got, tc.expectCC) {
				t.Fatalf("expected cc %v but got %v", tc.expectCC, got)
			}

			if lastEmail.ExtractTplID() != email.TplAddPolicyCoverage {
				t.Fatalf("expected tpl %v but got %v", email.TplAddPolicyCoverage, lastEmail.ExtractTplID())
			}

			var sent handlers.AddPolicyCoverageReq
			if err := json.Unmarshal([]byte(tc.body), &sent); err != nil {
				t.Fatalf("test body is not valid json: %v", err)
			}
			if string(lastEmail.ExtractMessage()) != string(sent.Message) {
				t.Fatalf("expected message %s but got %s", sent.Message, lastEmail.ExtractMessage())
			}
		})
	}
}

// TestInvalidMethodNeverSends guards a regression where a non-POST request
// was answered with 405 but the handler kept executing and could still send.
func TestInvalidMethodNeverSends(t *testing.T) {

	t.Parallel()

	body := `{"email_to":"holder@bar.com","message":{"foo":"bar"}}`

	routes := map[string]func(email.MailProvider) http.HandlerFunc{
		"add-policy-vehicle":  handlers.AddPolicyVehicle,
		"add-policy-driver":   handlers.AddPolicyDriver,
		"add-policy-address":  handlers.AddPolicyAddress,
		"add-policy-coverage": handlers.AddPolicyCoverage,
	}

	for name, handler := range routes {
		t.Run(name, func(t *testing.T) {
			testEmail := mockemail.NewClient()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/comms/"+name, strings.NewReader(body))

			handler(testEmail)(w, req)

			if w.Result().StatusCode != http.StatusMethodNotAllowed {
				t.Fatalf("expected 405, got %d", w.Result().StatusCode)
			}
			if !testEmail.SendLogs().IsEmpty() {
				t.Fatalf("handler sent %d email(s) on a rejected method", len(testEmail.SendLogs()))
			}
		})
	}
}

// TestMissingRecipientRejected covers the shared recipient validation on the
// pre-existing routes.
func TestMissingRecipientRejected(t *testing.T) {

	t.Parallel()

	routes := map[string]func(email.MailProvider) http.HandlerFunc{
		"add-policy-vehicle": handlers.AddPolicyVehicle,
		"add-policy-driver":  handlers.AddPolicyDriver,
		"add-policy-address": handlers.AddPolicyAddress,
	}

	for name, handler := range routes {
		t.Run(name, func(t *testing.T) {
			testEmail := mockemail.NewClient()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/comms/"+name, strings.NewReader(`{"email_to":"  ","message":{}}`))

			handler(testEmail)(w, req)

			if w.Result().StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Result().StatusCode)
			}
			if !testEmail.SendLogs().IsEmpty() {
				t.Fatal("email was sent without a recipient")
			}
		})
	}
}
