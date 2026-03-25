package libdns_regru

import (
	"context"

	"github.com/libdns/libdns"
)

// Requests

type RemoveDomainRequest struct {
	Domains       []GeneralZoneRequest `json:"domains"`
	Subdomain     string               `json:"subdomain"`
	RecordType    string               `json:"record_type"`
	Priority      *uint                `json:"priority,omitempty"`
	ContentFilter string               `json:"content,omitempty"`
}

// Responses:

type RemoveResponse = APIResponse[RemoveDomainsAnswer]

type RemoveDomainsAnswer struct {
	Domains []RemoveDomainResponse `json:"domains"`
}

type RemoveDomainResponse struct {
	GeneralResponseErrorInfoAndResult
	DName     string `json:"dname"`
	ServiceID string `json:"service_id,omitempty"`
}

func (self *RegruClient) RemoveZoneRecord(ctx context.Context, zone string, record libdns.Record) (*RemoveDomainResponse, error) {
	var respBody RemoveDomainResponse
	_, err := self.Client.R().SetBody(RemoveDomainRequest{
		Domains: []GeneralZoneRequest{{
			DName: zone,
		}},
		Subdomain:  record.RR().Name,
		RecordType: record.RR().Type,
		// we don't use other fields anyway
	}).
		SetContext(ctx).
		SetResult(&respBody).
		Post("/zone/remove_record")
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}
