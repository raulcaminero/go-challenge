package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gloveboxhq/glovebox-go-code-challenge/comms/email"
)

func AddPolicyVehicle(emailsvc email.MailProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		payload := AddPolicyVehicleReq{}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if !requireRecipient(w, payload.EmailTo) {
			return
		}

		if err := emailsvc.Send([]string{payload.EmailTo}, nil, payload.Message, email.TplAddPolicyVehicle); err != nil {
			http.Error(w, fmt.Sprintf("error sending email: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func AddPolicyDriver(emailsvc email.MailProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		payload := AddPolicyDriverReq{}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if !requireRecipient(w, payload.EmailTo) {
			return
		}

		if err := emailsvc.Send([]string{payload.EmailTo}, nil, payload.Message, email.TplAddPolicyDriver); err != nil {
			http.Error(w, fmt.Sprintf("error sending email: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func AddPolicyAddress(emailsvc email.MailProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		payload := AddPolicyAddressReq{}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if !requireRecipient(w, payload.EmailTo) {
			return
		}

		if err := emailsvc.Send([]string{payload.EmailTo}, nil, payload.Message, email.TplAddPolicyAddress); err != nil {
			http.Error(w, fmt.Sprintf("error sending email: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

// AddPolicyCoverage sends the add-policy-coverage notification to the primary
// recipient and carbon-copies everyone listed in email_cc, as one email.
func AddPolicyCoverage(emailsvc email.MailProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		payload := AddPolicyCoverageReq{}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if !requireRecipient(w, payload.EmailTo) {
			return
		}

		cc := normalizeCC(payload.EmailTo, payload.EmailCC)

		if err := emailsvc.Send([]string{payload.EmailTo}, cc, payload.Message, email.TplAddPolicyCoverage); err != nil {
			http.Error(w, fmt.Sprintf("error sending email: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

// requireRecipient rejects a request with no primary recipient. It writes the
// 400 response itself and reports whether the handler may continue.
func requireRecipient(w http.ResponseWriter, emailTo string) bool {
	if strings.TrimSpace(emailTo) == "" {
		http.Error(w, "email_to is required", http.StatusBadRequest)
		return false
	}
	return true
}

// normalizeCC trims whitespace, drops blanks and duplicates, and removes the
// primary recipient from the CC list so nobody receives the email twice.
// Comparison is case-insensitive because mailbox addresses are treated that
// way by every mainstream provider. It returns nil when nothing remains.
func normalizeCC(to string, cc []string) []string {
	seen := map[string]struct{}{strings.ToLower(strings.TrimSpace(to)): {}}
	out := make([]string, 0, len(cc))
	for _, addr := range cc {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		key := strings.ToLower(addr)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, addr)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
