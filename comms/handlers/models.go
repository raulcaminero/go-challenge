package handlers

import (
	"encoding/json"
)

type AddPolicyVehicleReq struct {
	EmailTo string          `json:"email_to"`
	Message json.RawMessage `json:"message"`
}

type AddPolicyDriverReq struct {
	EmailTo string          `json:"email_to"`
	Message json.RawMessage `json:"message"`
}

type AddPolicyAddressReq struct {
	EmailTo string          `json:"email_to"`
	Message json.RawMessage `json:"message"`
}

// AddPolicyCoverageReq follows the same contract as the other comms
// operations and additionally accepts carbon-copy recipients.
type AddPolicyCoverageReq struct {
	EmailTo string          `json:"email_to"`
	EmailCC []string        `json:"email_cc"`
	Message json.RawMessage `json:"message"`
}
