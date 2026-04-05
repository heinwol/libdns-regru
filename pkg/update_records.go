package libdns_regru

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/libdns/libdns"
)

// Requests

type UpdateRecordsRequest struct {
	Domains []UpdateZoneRequest `json:"domains"`
}

func (self UpdateRecordsRequest) getCommandName() string {
	return "update_records"
}

type UpdateZoneRequest struct {
	DName      string         `json:"dname"`
	ActionList []UpdateAction `json:"action_list"`
}

type UpdateAction struct {
	ActionName string `json:"action"`
	requestWithName
}

// Responses:

type UpdateResponse = APIResponse[UpdateDomainsAnswer]

type UpdateDomainsAnswer struct {
	Domains []UpdateZoneResponse `json:"domains"`
}

type UpdateZoneResponse struct {
	GeneralResponseErrorInfoAndResult
	DName      string                 `json:"dname"`
	ServiceID  json.Number            `json:"service_id,omitempty"`
	ActionList []UpdateActionResponse `json:"action_list"`
}

type UpdateActionResponse struct {
	GeneralResponseErrorInfoAndResult
	ActionName string `json:"action"`
}

func (self *RegruClient) UpdateZoneRecords(
	ctx context.Context,
	zone Zone,
	records []libdns.Record,
) (*UpdateZoneResponse, error) {
	domain := UpdateZoneRequest{
		DName: zone,
	}
	for _, record := range records {
		record_converted, err := addRequestFromLibdns(zone, record)
		if err != nil {
			return nil, err
		}
		domain.ActionList = append(domain.ActionList, UpdateAction{
			ActionName:      record_converted.getCommandName(),
			requestWithName: record_converted,
		})
	}

	req := UpdateRecordsRequest{
		Domains: []UpdateZoneRequest{domain},
	}

	var respBody UpdateResponse
	_, err := self.Client.R().
		SetBody(req).
		SetContext(ctx).
		SetResult(&respBody).
		Post(getUrl(req))
	if err != nil {
		return nil, err
	}
	if len(respBody.Answer.Domains) == 0 {
		return nil, fmt.Errorf("no domains match zone `%s`", zone)
	}
	if len(respBody.Answer.Domains) > 1 {
		slog.Warn("zone matched several domains, taking the first one:", "Domains", respBody.Answer.Domains)
	}
	return &respBody.Answer.Domains[0], err
}

// Returns records that were successfully altered; if there was any failure, returns it as well
func AnalyzeUpdateResponse(
	resp *UpdateZoneResponse,
	zone Zone,
	records []libdns.Record,
) ([]libdns.Record, error) {
	result := []libdns.Record{}
	errors_encountered := []error{}

	for _, record := range records {
		found, err := wasRecordSuccessfullyUpdated(*resp, zone, record)
		if err != nil {
			errors_encountered = append(errors_encountered, err)
		} else if found {
			result = append(result, record)
		} else {
			errors_encountered = append(
				errors_encountered,
				fmt.Errorf("record %+v was not found in answer", record),
			)
		}
	}
	return result, errors.Join(errors_encountered...)
}

// Returns `true` if corresponding record was found in answer, `false` if not found and an error
// if the record is unsupported or, finally, the request itself resulted in an error
func wasRecordSuccessfullyUpdated(
	resp UpdateZoneResponse,
	zone Zone,
	record libdns.Record,
) (bool, error) {

	record_converted, err := addRequestFromLibdns(zone, record)
	if err != nil {
		return false, err
	}
	if resp.DName != zone {
		return false, fmt.Errorf("incorrect zone for comparison: %s", zone)
	}
	for _, action_resp := range resp.ActionList {
		if action_resp.ActionName == record_converted.getCommandName() {
			if x := action_resp.intoError(); x != nil {
				return false, x
			}
			return true, nil
		}
	}
	return false, nil
}
