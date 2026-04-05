package libdns_regru

import (
	"context"
	"encoding/json"
)

// Requests:

type GetDomainRequest struct {
	DName string `json:"dname"`
}

type GetResourceRecordsRequest struct {
	Domains []GetDomainRequest `json:"domains"`
}

// Responses:

type GetDomainResponse struct {
	GeneralResponseErrorInfoAndResult
	DName     string      `json:"dname"`
	Records   []DNSRecord `json:"rrs"`
	ServiceID json.Number `json:"service_id,omitempty"`
	ServType  string      `json:"servtype,omitempty"`
	SOA       SOA         `json:"soa"`
}

func (self GetDomainResponse) getDName() string {
	return self.DName
}

type SOA struct {
	MinimumTTL string `json:"minimum_ttl,omitempty"`
	TTL        string `json:"ttl,omitempty"`
}

type GetDomainsAnswer struct {
	Domains []GetDomainResponse `json:"domains"`
}

type GetResourceRecordsResponse = APIResponse[GetDomainsAnswer]

func (self *RegruClient) GetZoneRecords(ctx context.Context, zone Zone) (*GetDomainResponse, error) {
	var respBody GetResourceRecordsResponse
	_, err := self.Client.R().SetBody(GetResourceRecordsRequest{
		Domains: []GetDomainRequest{{
			DName: zone,
		}},
	}).
		SetContext(ctx).
		SetResult(&respBody).
		Post("/zone/get_resource_records")

	if err != nil {
		return nil, err
	}
	return searchZoneInAnswerDomain(respBody.Answer.Domains, zone)
}
