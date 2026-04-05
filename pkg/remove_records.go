package libdns_regru

import (
	"context"
	"encoding/json"

	"github.com/libdns/libdns"
)

// Requests

type RemoveRecordRequest struct {
	Domains       []GeneralZoneRequest `json:"domains"`
	Subdomain     string               `json:"subdomain"`
	RecordType    string               `json:"record_type"`
	Priority      *uint                `json:"priority,omitempty"`
	ContentFilter string               `json:"content,omitempty"`
}

func (self RemoveRecordRequest) getCommandName() string {
	return "remove_record"
}

// Responses:

type RemoveResponse = APIResponse[RemoveRecordAnswer]

type RemoveRecordAnswer struct {
	Domains []RemoveRecordResponse `json:"domains"`
}

type RemoveRecordResponse struct {
	GeneralResponseErrorInfoAndResult
	DName     string      `json:"dname"`
	ServiceID json.Number `json:"service_id,omitempty"`
}

// Removes a record for zone. The total conversion between [libdns.Record] and [DNSRecord] is not
// performed, just `Name` and `Type` fields are used in the request.
func (self *RegruClient) RemoveZoneRecord(
	ctx context.Context,
	zone Zone,
	record libdns.Record,
) (*RemoveResponse, error) {
	req := RemoveRecordRequest{
		Domains: []GeneralZoneRequest{{
			DName: zone,
		}},
		Subdomain:  record.RR().Name,
		RecordType: record.RR().Type,
		// we don't use other fields anyway
	}
	var respBody RemoveResponse
	_, err := self.Client.R().
		SetBody(req).
		SetContext(ctx).
		SetResult(&respBody).
		Post(getUrl(req))
	return &respBody, err
}
