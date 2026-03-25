package libdns_regru

import (
	"encoding/json"
)

// type APIResponseError struct {

// }

type APIResponse[T any] struct {
	GeneralResponseErrorInfoAndResult
	Answer       T      `json:"answer,omitempty"`
	CharSet      string `json:"charset,omitempty"`
	MessageStore string `json:"messagestore,omitempty"`
}

type GeneralResponseErrorInfoAndResult struct {
	ErrorCode   string `json:"error_code,omitempty"`
	ErrorText   string `json:"error_text,omitempty"`
	ErrorParams any    `json:"error_params,omitempty"`
	Result      string `json:"result"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DNSRecordWOPriority struct {
	Rectype  string      `json:"rectype"`
	Subname  string      `json:"subname"`
	Content  string      `json:"content"`
	Priority json.Number `json:"priority,omitempty"`
	State    string      `json:"state,omitempty"`
}

type DNSRecord struct {
	Rectype  string `json:"rectype"`
	Subname  string `json:"subname"`
	Content  string `json:"content"`
	Priority *uint16
	State    string `json:"state,omitempty"`
}

// Priority behavior is inconsistent between requests and responses, therefore we need
// custom unmarshaling
func (r *DNSRecord) UnmarshalJSON(data []byte) error {
	type alias DNSRecord
	var aux struct {
		alias
		// for requests (and sometimes responses)
		Priority *json.Number `json:"priority"`
		// for responses
		Prio *json.Number `json:"prio"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	*r = DNSRecord(aux.alias)

	n := aux.Priority
	if n == nil {
		n = aux.Prio
	}
	if n != nil {
		v, err := n.Int64()
		if err != nil {
			return err
		}
		vv := uint16(v)
		r.Priority = &vv
	}

	return nil
}

// Priority behavior is inconsistent between requests and responses, therefore we need
// custom marshaling. Requests always use the `priority` variant
func (r DNSRecord) MarshalJSON() ([]byte, error) {
	type alias DNSRecord
	var aux struct {
		alias
		Priority uint16 `json:"priority,omitempty"`
	}
	aux.alias = alias(r)
	aux.Priority = *r.Priority
	return json.Marshal(aux)
}

type GeneralZoneRequest struct {
	DName string `json:"dname"`
}
