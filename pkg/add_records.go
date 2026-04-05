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

func (self AddAliasRequest) getCommandName() string {
	return "add_alias"
}

type AddAAAARequest struct {
	GeneralAddDomainRequest
	IPAddr string `json:"ipaddr"`
}

func (self AddAAAARequest) getCommandName() string {
	return "add_aaaa"
}

type AddCAARequest struct {
	GeneralAddDomainRequest
	Flags uint8  `json:"flags"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

func (self AddCAARequest) getCommandName() string {
	return "add_caa"
}

type AddCNAMERequest struct {
	GeneralAddDomainRequest
	Canonical string `json:"canonical_name"`
}

func (self AddCNAMERequest) getCommandName() string {
	return "add_cname"
}

type AddMXRequest struct {
	GeneralAddDomainRequest
	MailServer string  `json:"mail_server"`
	Priority   *uint16 `json:"priority,omitempty"`
}

func (self AddMXRequest) getCommandName() string {
	return "add_mx"
}

type AddNSRequest struct {
	GeneralAddDomainRequest
	DNSServer    string `json:"dns_server"`
	RecordNumber *int   `json:"record_number,omitempty"`
}

func (self AddNSRequest) getCommandName() string {
	return "add_ns"
}

type AddTXTRequest struct {
	GeneralAddDomainRequest
	Text string `json:"text"`
}

func (self AddTXTRequest) getCommandName() string {
	return "add_txt"
}

// Responses:

type AddResponse = APIResponse[DomainsAnswer]

// type AddResponse = APIResponse[AddDomainsAnswer]
// type AddDomainsAnswer struct {
// 	Domains []AddDomainResponse `json:"domains"`
// }

// type AddDomainResponse struct {
// 	GeneralResponseErrorInfoAndResult
// 	DName     string      `json:"dname"`
// 	ServiceID json.Number `json:"service_id,omitempty"`
// }

func (self *RegruClient) AddZoneRecord(
	ctx context.Context,
	zone Zone,
	record libdns.Record,
) (*SimpleDomainResponse, error) {
	req, err := addRequestFromLibdns(zone, record)
	if err != nil {
		return nil, err
	}
	var respBody AddResponse
	_, err = self.Client.R().
		SetBody(req).
		SetContext(ctx).
		SetResult(&respBody).
		Post(getUrl(req))
	if err != nil {
		return nil, err
	}
	return searchZoneInAnswerDomain(respBody.Answer.Domains, zone)
}

func addRequestFromLibdns(
	zone Zone,
	record libdns.Record,
) (requestWithName, error) {
	basic_req := GeneralAddDomainRequest{
		Domains: []GeneralZoneRequest{{
			DName: zone,
		}},
		Subdomain: record.RR().Name,
	}

	var specific_request requestWithName

	switch rec_t := record.(type) {
	case libdns.Address:
		if rec_t.IP.Is4() {
			specific_request = AddAliasRequest{
				GeneralAddDomainRequest: basic_req,
				IPAddr:                  rec_t.IP.String(),
			}
		} else {
			specific_request = AddAAAARequest{
				GeneralAddDomainRequest: basic_req,
				IPAddr:                  rec_t.IP.String(),
			}
		}
	case libdns.CAA:
		specific_request = AddCAARequest{
			GeneralAddDomainRequest: basic_req,
			Flags:                   rec_t.Flags,
			Tag:                     rec_t.Tag,
			Value:                   rec_t.Value,
		}
	case libdns.CNAME:
		specific_request = AddCNAMERequest{
			GeneralAddDomainRequest: basic_req,
			Canonical:               rec_t.Target,
		}
	case libdns.MX:
		specific_request = AddMXRequest{
			GeneralAddDomainRequest: basic_req,
			MailServer:              rec_t.Target,
			Priority:                &rec_t.Preference,
		}
	case libdns.NS:
		specific_request = AddNSRequest{
			GeneralAddDomainRequest: basic_req,
			DNSServer:               rec_t.Target,
		}
	case libdns.TXT:
		specific_request = AddTXTRequest{
			GeneralAddDomainRequest: basic_req,
			Text:                    rec_t.Text,
		}
	default:
		return nil, fmt.Errorf("unsupported record type: %s", rec_t.RR().Type)
	}
	return specific_request, nil
}
