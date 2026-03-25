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

// TODO: Providers must not require additional provisioning steps by the callers; it
// should work simply by populating a struct and calling methods on it. If your DNS
// service requires long-lived state or some extra provisioning step, do it implicitly
// when methods are called; sync.Once can help with this, and/or you can use a
// sync.(RW)Mutex in your Provider struct to synchronize implicit provisioning.

// Provider facilitates DNS record manipulation with <TODO: PROVIDER NAME>.
type Provider struct {
	// TODO: Put config fields here (with snake_case json struct tags on exported fields), for example:
	Username string `json:"username"`
	Password string `json:"password"`

	client     *RegruClient
	once       sync.Once
	init_error error
	// Exported config fields should be JSON-serializable or omitted (`json:"-"`)
}

func (p *Provider) initClient(ctx context.Context) error {
	p.once.Do(func() {
		p.client, p.init_error = NewRegruClient(Credentials{
			Username: p.Username,
			Password: p.Password,
		})
		if p.init_error == nil {
			p.client.Client.SetContext(ctx)
		}
	})
	return p.init_error
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	err := p.initClient(ctx)
	if err != nil {
		return nil, err
	}

	records, err := p.client.GetZoneRecords(ctx, zone)
	if err != nil {
		return nil, err
	}
	records_conv, err := records.IntoLibnsRecords()
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
