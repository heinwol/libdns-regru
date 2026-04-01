// Package libdnstemplate implements a DNS record management client compatible
// with the libdns interfaces for <PROVIDER NAME>. TODO: This package is a
// template only. Customize all godocs for actual implementation.
package libdns_regru

import (
	"context"
	"fmt"
	"sync"

	"github.com/libdns/libdns"
)

type Zone = string

// Provider facilitates DNS record manipulation with <TODO: PROVIDER NAME>.
type Provider struct {
	Username string `json:"username"`
	Password string `json:"password"`

	client onceCell[RegruClient]
	soa    sync.Map // map[Zone]SOA
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
	// e := libdns.AtomicErr
	return records_conv, err
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	// Make sure to return RR-type-specific structs, not libdns.RR structs.
	return nil, fmt.Errorf("TODO: not implemented")
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	// Make sure to return RR-type-specific structs, not libdns.RR structs.
	return nil, fmt.Errorf("TODO: not implemented")
}

// DeleteRecords deletes the specified records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	// Make sure to return RR-type-specific structs, not libdns.RR structs.
	return nil, fmt.Errorf("TODO: not implemented")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
