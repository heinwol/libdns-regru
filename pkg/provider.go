// Package libdnstemplate implements a DNS record management client compatible
// with the libdns interfaces for <PROVIDER NAME>. TODO: This package is a
// template only. Customize all godocs for actual implementation.
package libdns_regru

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"slices"
	"sync"

	"github.com/libdns/libdns"
)

type Zone = string

// Provider facilitates DNS record manipulation with <TODO: PROVIDER NAME>.
type Provider struct {
	Username string `json:"username"`
	Password string `json:"password"`

	client    onceCell[RegruClient]
	soa_cache sync.Map // map[Zone]SOA
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	err := p.initClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Inner.GetZoneRecords(ctx, zone)
	if err != nil {
		return nil, err
	}
	records_conv, err := resp.IntoLibdnsRecords()
	return records_conv, err
}

// AppendRecords adds records to the zone. It returns the records that were added.
// Beware, though: this method would not change TTL (that's reflected in the output records).
// If you want to change TTL, use [Provider.SetRecords]
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	err := p.initClient(ctx)
	if err != nil {
		return nil, err
	}

	currentSOA, err := p.getSOA(ctx, zone)
	if err != nil {
		return nil, err
	}
	currentTTL, err := fromRegruTTL(currentSOA.TTL)
	if err != nil {
		return nil, err
	}

	recordsWithAlteredTTL, changed, err1 := changeTTLInLibdnsRecords(records, currentTTL, currentTTL)
	if changed {
		slog.Warn("attempt to change TTL while appending records; that's unsupported")
	}
	appendedRecords := []libdns.Record{}

	for _, record := range recordsWithAlteredTTL {
		_, err := p.client.Inner.AddZoneRecord(ctx, zone, record)
		if err != nil {
			return appendedRecords, errors.Join(err1, err)
		}
		appendedRecords = append(appendedRecords, record)
	}
	return appendedRecords, err1
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	err := p.initClient(ctx)
	if err != nil {
		return nil, err
	}

	currentSOA, err := p.getSOA(ctx, zone)
	if err != nil {
		return nil, err
	}
	currentTTL, err := fromRegruTTL(currentSOA.TTL)
	if err != nil {
		return nil, err
	}

	changeTTL := slices.MinFunc(records, func(r1, r2 libdns.Record) int {
		return cmp.Compare(r1.RR().TTL, r2.RR().TTL)
	}).RR().TTL
	recordsWithAlteredTTL, changed, errTTLChangeInRecords := changeTTLInLibdnsRecords(records, currentTTL, changeTTL)

	if !changed && errTTLChangeInRecords == nil {
		recordsWithAlteredTTL = records
	}
	var errSOAUpdate error
	if changed {
		errSOAUpdate = p.updateSOA(ctx, zone, changeTTL)
	}

	resp, err := p.client.Inner.UpdateZoneRecords(ctx, zone, recordsWithAlteredTTL)
	if err != nil {
		return nil, err
	}
	resultRecords, err := AnalyzeUpdateResponse(resp, zone, recordsWithAlteredTTL)
	return resultRecords, errors.Join(err, errTTLChangeInRecords, errSOAUpdate)
}

// DeleteRecords deletes the specified records from the zone. It returns the records that were deleted.
// Each record results in a separate request. If an error occurs mid-way, no attempts to recover
// or undo previous deletions is made; it's up to the user to handle.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	err := p.initClient(ctx)
	if err != nil {
		return nil, err
	}

	deleted_records := []libdns.Record{}
	for _, record := range records {
		_, err := p.client.Inner.RemoveZoneRecord(ctx, zone, record)
		if err != nil {
			return deleted_records, err
		}
		deleted_records = append(deleted_records, record)
	}
	return deleted_records, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
