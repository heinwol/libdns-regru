package libdns_regru

import (
	"context"
	"fmt"

	"github.com/libdns/libdns"
)

// Requests

type GeneralAddDomainRequest struct {
	Domains   []GeneralZoneRequest `json:"domains"`
	Subdomain string               `json:"subdomain"`
}

//

type AddAliasRequest struct {
	GeneralAddDomainRequest
	IPAddr string `json:"ipaddr"`
}

// type AddAAAARequest struct {
// 	GeneralAddDomainRequest
// 	IPAddr string `json:"ipaddr"`
// }

type AddCAARequest struct {
	GeneralAddDomainRequest
	Flags uint8  `json:"flags"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

type AddCNAMERequest struct {
	GeneralAddDomainRequest
	Canonical string `json:"canonical_name"`
}

type AddMXRequest struct {
	GeneralAddDomainRequest
	MailServer string  `json:"mail_server"`
	Priority   *uint16 `json:"priority,omitempty"`
}

type AddNSRequest struct {
	GeneralAddDomainRequest
	DNSServer    string `json:"dns_server"`
	RecordNumber *int   `json:"record_number,omitempty"`
}

type AddTXTRequest struct {
	GeneralAddDomainRequest
	Text string `json:"text"`
}

// Responses:

type AddResponse = APIResponse[AddDomainsAnswer]

type AddDomainsAnswer struct {
	Domains []AddDomainResponse `json:"domains"`
}

type AddDomainResponse struct {
	GeneralResponseErrorInfoAndResult
	DName     string `json:"dname"`
	ServiceID string `json:"service_id,omitempty"`
}

func (self *RegruClient) AddZoneRecord(
	ctx context.Context,
	zone string,
	record libdns.Record,
) (*AddDomainResponse, error) {
	basic_req := GeneralAddDomainRequest{
		Domains: []GeneralZoneRequest{{
			DName: zone,
		}},
		Subdomain: record.RR().Name,
	}

	var specific_request any
	var req_url string

	switch rec_t := record.(type) {
	case libdns.Address:
		specific_request = AddAliasRequest{
			GeneralAddDomainRequest: basic_req,
			IPAddr:                  rec_t.IP.String(),
		}
		if rec_t.IP.Is4() {
			req_url = "/zone/add_alias"
		} else {
			req_url = "/zone/add_aaaa"

		}
	case libdns.CNAME:
		specific_request = AddCNAMERequest{
			GeneralAddDomainRequest: basic_req,
			Canonical:               rec_t.Target,
		}
		req_url = "/zone/add_cname"
	case libdns.MX:
		specific_request = AddMXRequest{
			GeneralAddDomainRequest: basic_req,
			MailServer:              rec_t.Target,
			Priority:                &rec_t.Preference,
		}
		req_url = "/zone/add_mx"
	case libdns.NS:
		specific_request = AddNSRequest{
			GeneralAddDomainRequest: basic_req,
			DNSServer:               rec_t.Target,
		}
		req_url = "/zone/add_ns"
	case libdns.TXT:
		specific_request = AddTXTRequest{
			GeneralAddDomainRequest: basic_req,
			Text:                    rec_t.Text,
		}
		req_url = "/zone/add_txt"
	default:
		return nil, fmt.Errorf("unsupported record type: %s", rec_t.RR().Type)
	}
	var respBody AddDomainResponse
	_, err := self.Client.R().
		SetBody(specific_request).
		SetContext(ctx).
		SetResult(&respBody).
		Post(req_url)
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}

// func (self *RegruClient) commonAddZoneRecord(
// 	ctx context.Context,

// 	zone string,
// 	record libdns.Record,
// 	add_ttl bool,
// ) (*AddDomainResponse, error) {
// 	basic_req := GeneralAddDomainRequest{
// 		Domains: []GeneralZoneRequest{{
// 			DName: zone,
// 		}},
// 		Subdomain: record.RR().Name,
// 	}
// 	// specific_request :=
// 	var respBody AddDomainResponse
// }
