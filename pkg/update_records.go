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

// We need this to flatten the struct
func (a UpdateAction) MarshalJSON() ([]byte, error) {
	// marshal the inner request first
	inner, err := json.Marshal(a.requestWithName)
	if err != nil {
		return nil, err
	}

	// merge action name into it
	var m map[string]any
	if err := json.Unmarshal(inner, &m); err != nil {
		return nil, err
	}
	m["action"] = a.ActionName
	return json.Marshal(m)
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
	// first we try to match records 1:1 to action list, but there's no guarantee from the server
	// it would work
	if len(resp.ActionList) == len(records) {
		result := []libdns.Record{}
		errors_encountered := []error{}
		for i := range len(records) {
			record_converted, err := addRequestFromLibdns(zone, records[i])
			if err != nil {
				errors_encountered = append(errors_encountered, err)
				continue
			}
			if resp.ActionList[i].ActionName == record_converted.getCommandName() {
				// here we just assume the corresponding action is the same as in record
				if x := resp.ActionList[i].intoError(); x != nil {
					errors_encountered = append(
						errors_encountered,
						fmt.Errorf("record %+v was not found in answer", records[i]),
					)
				} else {
					result = append(result, records[i])
				}
			} else {
				// here we allow some false positives, but it's our best bet
				found, err := wasRecordSuccessfullyUpdated(*resp, zone, records[i])
				if err != nil {
					errors_encountered = append(errors_encountered, err)
					continue
				}
				if found {
					result = append(result, records[i])
				} else {
					errors_encountered = append(
						errors_encountered,
						fmt.Errorf("record %+v was not found in answer", records[i]),
					)
				}
			}
		}
		return result, errors.Join(errors_encountered...)
	} else {
		return analyzeUpdateResponseFallback(resp, zone, records)
	}
}

func analyzeUpdateResponseFallback(
	resp *UpdateZoneResponse,
	zone Zone,
	records []libdns.Record,
) ([]libdns.Record, error) {
	slog.Warn(
		"resorting to fallback when trying to decide what records were updated, expect some false positives",
	)
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
// if the record is unsupported or, finally, the request itself resulted in an error.
//
// Warning: It is a fallback measure, as it would return incorrect results if there were several
// requests of the same type (all false-positives)
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
