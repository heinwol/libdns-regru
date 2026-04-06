// Package libdnstemplate implements a DNS record management client compatible
// with the libdns interfaces for <PROVIDER NAME>. TODO: This package is a
// template only. Customize all godocs for actual implementation.
package libdns_regru

import (
	"context"
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
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	err := p.initClient(ctx)
	if err != nil {
		return nil, err
	}

	appended_records := []libdns.Record{}
	for _, record := range records {
		_, err := p.client.Inner.AddZoneRecord(ctx, zone, record)
		if err != nil {
			return appended_records, err
		}
		appended_records = append(appended_records, record)
	}
	return appended_records, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	err := p.initClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Inner.UpdateZoneRecords(ctx, zone, records)
	if err != nil {
		return nil, err
	}
	return AnalyzeUpdateResponse(resp, zone, records)
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
