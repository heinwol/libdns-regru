package libdns_regru

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

func NewProviderForTests(ctx context.Context) (*Provider, error) {
	client, err := NewRegruClientForTests(ctx)
	if err != nil {
		return nil, err
	}
	result := Provider{
		Password: client.Credentials.Password,
		Username: client.Credentials.Username,
	}
	_, err = result.Client.Do(func() (*RegruClient, error) {
		return client, nil
	})
	return &result, err

}

func (p *Provider) initClient(ctx context.Context) error {
	_, err := p.Client.Do(func() (*RegruClient, error) {
		if p.Username == "" {
			return nil, fmt.Errorf("regru: username is required")
		}
		if p.Password == "" {
			return nil, fmt.Errorf("regru: password is required")
		}
		client, err := NewRegruClient(Credentials{
			Username: p.Username,
			Password: p.Password,
		})
		if err == nil {
			client.Client.SetContext(ctx)
		}
		return client, err
	})
	return err
}

func (p *Provider) GetSOA(ctx context.Context, zone Zone) (*SOA, error) {
	err := p.initClient(ctx)
	if err != nil {
		return nil, err
	}

	soa, ok := p.soa_cache.Load(zone)
	if ok {
		soa_ := soa.(SOA)
		return &soa_, nil
	}

	resp, err := p.Client.Inner.GetZoneRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	if resp.SOA.TTL == "" {
		return &resp.SOA, fmt.Errorf("Server returned empty SOA")
	}

	p.soa_cache.Store(zone, resp.SOA)
	return &resp.SOA, nil
}

func (p *Provider) storeSOA(zone Zone, soa SOA) {
	p.soa_cache.Store(zone, soa)
}

func (p *Provider) updateTTLRemote(ctx context.Context, zone Zone, ttl time.Duration) error {
	err := p.initClient(ctx)
	if err != nil {
		return err
	}

	var minTTL string
	cached_soa, err := p.GetSOA(ctx, zone)
	if err != nil {
		slog.Warn("could not get current SOA, requesting update either way", "err", err)
		minTTL = ""
	} else {
		minTTL = cached_soa.MinimumTTL
	}

	_, err = p.Client.Inner.DoUpdateSOARequest(ctx, zone, ttl)
	if err != nil {
		return err
	}
	p.soa_cache.Store(zone, SOA{
		MinimumTTL: minTTL,
		TTL:        intoRegruTTLWithRoundingToSeconds(ttl),
	})
	return nil
}
