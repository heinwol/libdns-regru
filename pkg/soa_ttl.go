package libdns_regru

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/libdns/libdns"
)

// Request

type UpdateSOARequest struct {
	Domains []GeneralZoneRequest `json:"domains"`
	SOA
}

// Response

type UpdateSOAResponse = APIResponse[DomainsAnswer]

func (self *RegruClient) DoUpdateSOARequest(
	ctx context.Context,
	zone Zone,
	ttl time.Duration,
) (*SimpleDomainResponse, error) {
	req := UpdateSOARequest{
		Domains: []GeneralZoneRequest{{
			DName: zone,
		}},
		SOA: SOA{
			// TODO: check whether empty minimum ttl works
			TTL: intoRegruTTLWithRoundingToSeconds(ttl),
		},
	}

	var respBody UpdateSOAResponse
	_, err := self.Client.R().
		SetBody(req).
		SetContext(ctx).
		SetResult(&respBody).
		Post("/zone/update_soa")
	if err != nil {
		return nil, err
	}
	return searchZoneInAnswerDomain(respBody.Answer.Domains, zone)
}

// returns **modified** (i.e. with unified SOA for zone) records and `true` if changes were made,
// otherwise returns (copies of) unmodified records and false
func changeTTLInLibdnsRecords(
	records []libdns.Record,
	currentTTL, changeTTL time.Duration,
) ([]libdns.Record, bool, error) {
	if len(records) == 0 {
		return records, false, nil
	}

	results := []libdns.Record{}
	recordsChanged := false
	errorsEncountered := []error{}

	for _, record := range records {
		if changeTTL != currentTTL {
			recordsChanged = true
		}
		if record.RR().TTL == changeTTL {
			results = append(results, record)
		} else {
			slog.Warn("TTL is inconsistent:", "TTL to set", changeTTL, "TTL in record", record.RR().TTL)
			new, err := withTTL(record, changeTTL)
			if err != nil {
				errorsEncountered = append(errorsEncountered, err)
			} else {
				results = append(results, new)
			}
		}
	}

	return results, recordsChanged, errors.Join(errorsEncountered...)
}

func withTTL(rec libdns.Record, ttl time.Duration) (libdns.Record, error) {
	switch r := rec.(type) {
	case libdns.TXT:
		r.TTL = ttl
		return r, nil
	case libdns.Address:
		r.TTL = ttl
		return r, nil
	case libdns.CNAME:
		r.TTL = ttl
		return r, nil
	case libdns.MX:
		r.TTL = ttl
		return r, nil
	case libdns.NS:
		r.TTL = ttl
		return r, nil
	case libdns.SRV:
		r.TTL = ttl
		return r, nil
	case libdns.CAA:
		r.TTL = ttl
		return r, nil
	case libdns.RR:
		r.TTL = ttl
		return r, nil
	default:
		return nil, fmt.Errorf("unsupported record type %T", rec)
	}
}
