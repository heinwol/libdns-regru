package libdns_regru

import (
	"context"
	"time"
)

func (p *Provider) initClient(ctx context.Context) error {
	_, err := p.client.Do(func() (*RegruClient, error) {
		client, err := NewRegruClient(Credentials{
			Username: p.Username,
			Password: p.Password,
		})
		if err != nil {
			client.Client.SetContext(ctx)
		}
		return client, err
	})
	return err
}

func (p *Provider) getSOA(ctx context.Context, zone Zone) (*SOA, error) {
	err := p.initClient(ctx)
	if err != nil {
		return nil, err
	}

	soa, ok := p.soa_cache.Load(zone)
	if ok {
		soa_ := soa.(SOA)
		return &soa_, nil
	}

	resp, err := p.client.Inner.GetZoneRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	p.soa_cache.Store(zone, resp.SOA)
	return &resp.SOA, nil
}

func (p *Provider) updateSOA(ctx context.Context, zone Zone, ttl time.Duration) error {
	err := p.initClient(ctx)
	if err != nil {
		return err
	}

	cached_soa, err := p.getSOA(ctx, zone)

	_, err = p.client.Inner.UpdateSOA(ctx, zone, ttl)
	if err != nil {
		return err
	}
	p.soa_cache.Store(zone, SOA{
		MinimumTTL: cached_soa.MinimumTTL,
		TTL:        intoRegruTTLWithRoundingToSeconds(ttl),
	})
	return nil
}
