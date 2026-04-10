package libdns_regru

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
)

// type APIResponseError struct {

// }

type APIResponse[T any] struct {
	GeneralResponseErrorInfoAndResult
	Answer       T      `json:"answer,omitempty"`
	CharSet      string `json:"charset,omitempty"`
	MessageStore string `json:"messagestore,omitempty"`
}

type APIResponseError struct {
	GeneralResponseErrorInfoAndResult
}

func (self *APIResponseError) Error() string {
	return fmt.Sprintf("reg.ru request resulted in error: %#v", self.GeneralResponseErrorInfoAndResult)
}

type SimpleDomainResponse struct {
	GeneralResponseErrorInfoAndResult
	DName     string      `json:"dname"`
	ServiceID json.Number `json:"service_id,omitempty"`
}

type DomainsAnswer struct {
	Domains []SimpleDomainResponse `json:"domains"`
}

type GeneralResponseErrorInfoAndResult struct {
	ErrorCode   string `json:"error_code,omitempty"`
	ErrorText   string `json:"error_text,omitempty"`
	ErrorParams any    `json:"error_params,omitempty"`
	Result      string `json:"result"`
}

func (self *GeneralResponseErrorInfoAndResult) intoError() *APIResponseError {
	if self.Result != "success" {
		return &APIResponseError{*self}
	} else {
		return nil
	}
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
	if r.Priority != nil {
		aux.Priority = *r.Priority
	}
	return json.Marshal(aux)
}

type GeneralZoneRequest struct {
	DName string `json:"dname"`
}

type requestWithName interface {
	getCommandName() string
}

func getUrl(req requestWithName) string {
	return fmt.Sprintf("/zone/%s", req.getCommandName())
}

type withDName interface {
	getDName() string
}

func (self SimpleDomainResponse) getDName() string {
	return self.DName
}

func searchZoneInAnswerDomain[DomainResponseT withDName](domains []DomainResponseT, zone Zone) (*DomainResponseT, error) {
	if len(domains) == 0 {
		return nil, fmt.Errorf("no domains match zone `%s`", zone)
	}
	if len(domains) > 1 {
		slog.Warn("zone matched several domains, searching for", "zone", zone, "in domains", domains)
		zone_index := slices.IndexFunc(
			domains,
			func(r DomainResponseT) bool {
				return r.getDName() == zone
			},
		)
		if zone_index == -1 {
			return nil, fmt.Errorf("could not find zone '%s' in response", zone)
		}
		return &domains[zone_index], nil
	}
	return &domains[0], nil
}
