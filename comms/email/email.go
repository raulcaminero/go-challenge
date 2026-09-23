package email

import "encoding/json"

type TplID string

const (
	TplAddPolicyVehicle  TplID = "add-policy-vehicle"
	TplAddPolicyDriver   TplID = "add-policy-driver"
	TplAddPolicyAddress  TplID = "add-policy-address"
	TplAddPolicyCoverage TplID = "add-policy-coverage"
)

// MailProvider delivers a templated message.
//
// to holds the primary recipients and cc the carbon-copy recipients; cc may be
// nil or empty for operations that only address To. A single Send is one
// email: every To and CC address receives the same message and can see each
// other, matching standard CC semantics.
type MailProvider interface {
	Send(to, cc []string, message json.RawMessage, tpl TplID) error
}
