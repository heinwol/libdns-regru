package libdns_regru

import (
	"fmt"
	"log/slog"
	"net/netip"
	"strconv"
	"time"

	"github.com/libdns/libdns"
)

func (self *DomainResponse) IntoLibnsRecords() ([]libdns.Record, error) {
	ttl, err := parseTTL(self.SOA.TTL)
	if err != nil {
		return nil, err
	}
	result := []libdns.Record{}
	for _, record := range self.Records {
		switch record.Rectype {
		case "TXT":
			result = append(result, libdns.TXT{
				Name: record.Subname,
				TTL:  ttl,
				Text: record.Content,
			})
		case "A", "AAAA":
			addr, err := netip.ParseAddr(record.Content)
			if err != nil {
				// return nil, err
				slog.Error("could not parse IP address:", "addr", record.Content)
				continue
			}
			result = append(result, libdns.Address{
				Name: record.Subname,
				TTL:  ttl,
				IP:   addr,
			})
		case "NS":
			result = append(result, libdns.NS{
				Name:   record.Subname,
				TTL:    ttl,
				Target: record.Content,
			})
		case "CNAME":
			result = append(result, libdns.CNAME{
				Name:   record.Subname,
				TTL:    ttl,
				Target: record.Content,
			})
		case "MX":
			pref, err := record.Prio.Int64()
			if err != nil {
				slog.Error("could not convert json Number to int:", "pref", record.Prio)
				continue
			}
			result = append(result, libdns.MX{
				Name:       record.Subname,
				TTL:        ttl,
				Preference: uint16(pref),
				Target:     record.Content,
			})
		default:
			slog.Error("unsupported record type:", "type", record.Rectype, "record", record)
			continue
		}

	}
	return result, nil
}

func parseTTL(s string) (time.Duration, error) {

	n, err := strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return 0, fmt.Errorf("bad ttl number: %w", err)
	}

	switch s[len(s)-1] {
	case 'h':
		return time.Duration(n) * time.Hour, nil
	case 'd':
		return time.Duration(n) * 24 * time.Hour, nil
	case 'w':
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	case 'm':
		slog.Warn("month TTL approximated as 30 days:", "ttl", s)
		return time.Duration(n) * 30 * 24 * time.Hour, nil
	default:
		secs, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("unsupported ttl unit in %q", s)
		}
		return time.Duration(secs) * time.Second, nil

	}
}
