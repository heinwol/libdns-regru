package libdns_regru

import (
	"fmt"
	"log/slog"
	"net/netip"
	"strconv"
	"time"

	"github.com/libdns/libdns"
)

func (self *GetDomainResponse) IntoLibdnsRecords() ([]libdns.Record, error) {
	ttl, err := fromRegruTTL(self.SOA.TTL)
	if err != nil {
		return nil, err
	}
	result := []libdns.Record{}
	for _, record := range self.Records {
		res, err := record.intoLibdnsRecordWithTTL(ttl)
		if err != nil {
			slog.Error("could not convert DNSRecord into regru.record:", "err", err)
			continue
		}
		result = append(result, res)
	}
	return result, nil
}

func fromLibdnsRecordWithTTL(record libdns.Record) (*DNSRecord, string, error) {
	result := DNSRecord{
		Rectype: record.RR().Type,
		Subname: record.RR().Name,
		// Content: record.RR().Data,
	}
	ttl := intoRegruTTLWithRoundingToSeconds(record.RR().TTL)
	switch rec_t := record.(type) {
	case libdns.Address:
		result.Content = rec_t.IP.String()
	case libdns.CNAME:
		result.Content = rec_t.Target
	case libdns.MX:
		result.Content = rec_t.Target
		result.Priority = &rec_t.Preference
	case libdns.NS:
		result.Content = rec_t.Target
	case libdns.TXT:
		result.Content = rec_t.Text
	default:
		return nil, ttl, fmt.Errorf("unsupported record type: %s", result.Rectype)
	}
	return &result, ttl, nil
}

func (self DNSRecord) intoLibdnsRecordWithTTL(ttl time.Duration) (libdns.Record, error) {

	switch self.Rectype {
	case "A", "AAAA":
		addr, err := netip.ParseAddr(self.Content)
		if err != nil {
			return nil, fmt.Errorf("could not parse IP address of %s: %w", self.Content, err)
		}
		return libdns.Address{
			Name: self.Subname,
			TTL:  ttl,
			IP:   addr,
		}, nil
	case "CNAME":
		return libdns.CNAME{
			Name:   self.Subname,
			TTL:    ttl,
			Target: self.Content,
		}, nil
	case "MX":
		return libdns.MX{
			Name:       self.Subname,
			TTL:        ttl,
			Preference: *self.Priority,
			Target:     self.Content,
		}, nil
	case "NS":
		return libdns.NS{
			Name:   self.Subname,
			TTL:    ttl,
			Target: self.Content,
		}, nil
	case "TXT":
		return libdns.TXT{
			Name: self.Subname,
			TTL:  ttl,
			Text: self.Content,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported record type: %s", self.Rectype)
	}
}

func intoRegruTTLWithRoundingToSeconds(ttl time.Duration) string {
	if ttl < 0 {
		slog.Error("TTL is somehow negative", "ttl", ttl)
	}
	if ttl%time.Second != 0 {
		slog.Warn("TTL will be rounded up to seconds", "ttl", ttl)
	}
	return strconv.FormatInt(int64(ttl/time.Second), 10)
}

func fromRegruTTL(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty TTL string")
	}

	n, err := strconv.Atoi(s)
	if err == nil {
		return time.Duration(n) * time.Second, nil
	}

	n, err = strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return 0, fmt.Errorf("bad ttl number: %w", err)
	}
	switch s[len(s)-1] {
	case 's':
		return time.Duration(n) * time.Second, nil
	case 'm':
		return time.Duration(n) * time.Minute, nil
	case 'h':
		return time.Duration(n) * time.Hour, nil
	case 'd':
		return time.Duration(n) * 24 * time.Hour, nil
	case 'w':
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported ttl unit in %q", s)
	}
}
