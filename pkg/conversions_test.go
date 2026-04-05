package libdns_regru

import (
	"net/netip"
	"testing"
	"time"

	"github.com/libdns/libdns"
)

// ---------------------------------------------------------------------------
// fromRegruTTL
// ---------------------------------------------------------------------------

func TestFromRegruTTL(t *testing.T) {
	cases := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		// numeric seconds (no suffix letter → falls through to the Atoi branch)
		{"3600", 3600 * time.Second, false},
		// NOTE: "0" triggers a known edge case — s[:len(s)-1] strips the last
		// digit ('0'), leaving "", which Atoi rejects before the fallthrough branch.
		// This test documents the current (arguably buggy) behavior.
		{"0", 0, true},
		// hour suffix
		{"1h", time.Hour, false},
		{"12h", 12 * time.Hour, false},
		// day suffix
		{"1d", 24 * time.Hour, false},
		{"7d", 7 * 24 * time.Hour, false},
		// week suffix
		{"2w", 2 * 7 * 24 * time.Hour, false},
		// month suffix (approximated)
		{"1m", 30 * 24 * time.Hour, false},
		// bad inputs
		{"xh", 0, true},  // non-numeric number part
		{"abc", 0, true}, // fully non-numeric (falls through, Atoi fails)
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			got, err := fromRegruTTL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("fromRegruTTL(%q): expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Errorf("fromRegruTTL(%q): unexpected error: %v", tc.input, err)
				return
			}
			if got != tc.want {
				t.Errorf("fromRegruTTL(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// intoRegruTTLWithRoundingToSeconds
// ---------------------------------------------------------------------------

func TestIntoRegruTTLWithRoundingToSeconds(t *testing.T) {
	cases := []struct {
		ttl  time.Duration
		want string
	}{
		{time.Hour, "3600"},
		{300 * time.Second, "300"},
		{0, "0"},
		// Sub-second fractions are truncated (logged as warning but still works)
		{1500 * time.Millisecond, "1"},
	}
	for _, tc := range cases {
		got := intoRegruTTLWithRoundingToSeconds(tc.ttl)
		if got != tc.want {
			t.Errorf("intoRegruTTLWithRoundingToSeconds(%v) = %q, want %q", tc.ttl, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// DNSRecord.intoLibdnsRecordWithTTL
// ---------------------------------------------------------------------------

func TestDNSRecordIntoLibdnsRecordWithTTL(t *testing.T) {
	ttl := time.Hour

	t.Run("A", func(t *testing.T) {
		rec := DNSRecord{Rectype: "A", Subname: "www", Content: "1.2.3.4"}
		got, err := rec.intoLibdnsRecordWithTTL(ttl)
		if err != nil {
			t.Fatal(err)
		}
		addr, ok := got.(libdns.Address)
		if !ok {
			t.Fatalf("expected libdns.Address, got %T", got)
		}
		if addr.IP.String() != "1.2.3.4" {
			t.Errorf("IP = %s, want 1.2.3.4", addr.IP)
		}
		if addr.Name != "www" {
			t.Errorf("Name = %s, want www", addr.Name)
		}
		if addr.TTL != ttl {
			t.Errorf("TTL = %v, want %v", addr.TTL, ttl)
		}
	})

	t.Run("AAAA", func(t *testing.T) {
		rec := DNSRecord{Rectype: "AAAA", Subname: "@", Content: "::1"}
		got, err := rec.intoLibdnsRecordWithTTL(ttl)
		if err != nil {
			t.Fatal(err)
		}
		addr, ok := got.(libdns.Address)
		if !ok {
			t.Fatalf("expected libdns.Address, got %T", got)
		}
		if !addr.IP.Is6() {
			t.Error("expected IPv6 address")
		}
	})

	t.Run("CNAME", func(t *testing.T) {
		rec := DNSRecord{Rectype: "CNAME", Subname: "alias", Content: "target.example.com."}
		got, err := rec.intoLibdnsRecordWithTTL(ttl)
		if err != nil {
			t.Fatal(err)
		}
		cname, ok := got.(libdns.CNAME)
		if !ok {
			t.Fatalf("expected libdns.CNAME, got %T", got)
		}
		if cname.Target != "target.example.com." {
			t.Errorf("Target = %s", cname.Target)
		}
	})

	t.Run("MX", func(t *testing.T) {
		prio := uint16(10)
		rec := DNSRecord{Rectype: "MX", Subname: "@", Content: "mail.example.com.", Priority: &prio}
		got, err := rec.intoLibdnsRecordWithTTL(ttl)
		if err != nil {
			t.Fatal(err)
		}
		mx, ok := got.(libdns.MX)
		if !ok {
			t.Fatalf("expected libdns.MX, got %T", got)
		}
		if mx.Preference != 10 {
			t.Errorf("Preference = %d, want 10", mx.Preference)
		}
		if mx.Target != "mail.example.com." {
			t.Errorf("Target = %s", mx.Target)
		}
	})

	t.Run("NS", func(t *testing.T) {
		rec := DNSRecord{Rectype: "NS", Subname: "@", Content: "ns1.example.com."}
		got, err := rec.intoLibdnsRecordWithTTL(ttl)
		if err != nil {
			t.Fatal(err)
		}
		ns, ok := got.(libdns.NS)
		if !ok {
			t.Fatalf("expected libdns.NS, got %T", got)
		}
		if ns.Target != "ns1.example.com." {
			t.Errorf("Target = %s", ns.Target)
		}
	})

	t.Run("TXT", func(t *testing.T) {
		rec := DNSRecord{Rectype: "TXT", Subname: "_acme-challenge", Content: "some-token"}
		got, err := rec.intoLibdnsRecordWithTTL(ttl)
		if err != nil {
			t.Fatal(err)
		}
		txt, ok := got.(libdns.TXT)
		if !ok {
			t.Fatalf("expected libdns.TXT, got %T", got)
		}
		if txt.Text != "some-token" {
			t.Errorf("Text = %s", txt.Text)
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		rec := DNSRecord{Rectype: "SRV", Subname: "_sip", Content: "whatever"}
		_, err := rec.intoLibdnsRecordWithTTL(ttl)
		if err == nil {
			t.Error("expected error for unsupported type SRV")
		}
	})

	t.Run("invalid_IP", func(t *testing.T) {
		rec := DNSRecord{Rectype: "A", Subname: "bad", Content: "not-an-ip"}
		_, err := rec.intoLibdnsRecordWithTTL(ttl)
		if err == nil {
			t.Error("expected error for invalid IP")
		}
	})
}

// ---------------------------------------------------------------------------
// fromLibdnsRecordWithTTL
// ---------------------------------------------------------------------------

func TestFromLibdnsRecordWithTTL(t *testing.T) {
	ttl := 300 * time.Second

	t.Run("Address_IPv4", func(t *testing.T) {
		rec := libdns.Address{Name: "www", TTL: ttl, IP: netip.MustParseAddr("1.2.3.4")}
		got, gotTTL, err := fromLibdnsRecordWithTTL(rec)
		if err != nil {
			t.Fatal(err)
		}
		if got.Content != "1.2.3.4" {
			t.Errorf("Content = %s", got.Content)
		}
		if gotTTL != "300" {
			t.Errorf("TTL string = %s, want 300", gotTTL)
		}
	})

	t.Run("TXT", func(t *testing.T) {
		rec := libdns.TXT{Name: "_acme", TTL: ttl, Text: "challenge-token"}
		got, _, err := fromLibdnsRecordWithTTL(rec)
		if err != nil {
			t.Fatal(err)
		}
		if got.Content != "challenge-token" {
			t.Errorf("Content = %s", got.Content)
		}
		if got.Rectype != "TXT" {
			t.Errorf("Rectype = %s, want TXT", got.Rectype)
		}
	})

	t.Run("MX", func(t *testing.T) {
		prio := uint16(20)
		rec := libdns.MX{Name: "@", TTL: ttl, Preference: prio, Target: "mail.example.com."}
		got, _, err := fromLibdnsRecordWithTTL(rec)
		if err != nil {
			t.Fatal(err)
		}
		if got.Priority == nil || *got.Priority != 20 {
			t.Errorf("Priority = %v, want 20", got.Priority)
		}
	})

	t.Run("unsupported_CAA", func(t *testing.T) {
		rec := libdns.CAA{Name: "@", TTL: ttl, Tag: "issue", Value: "letsencrypt.org"}
		_, _, err := fromLibdnsRecordWithTTL(rec)
		if err == nil {
			t.Error("expected error for CAA (not in fromLibdnsRecordWithTTL switch)")
		}
	})
}

// ---------------------------------------------------------------------------
// IntoLibdnsRecords (full pipeline)
// ---------------------------------------------------------------------------

func TestGetDomainResponseIntoLibdnsRecords(t *testing.T) {
	resp := GetDomainResponse{
		GeneralResponseErrorInfoAndResult: GeneralResponseErrorInfoAndResult{Result: "success"},
		DName:                             "example.com",
		SOA:                               SOA{TTL: "1h"},
		Records: []DNSRecord{
			{Rectype: "A", Subname: "www", Content: "1.2.3.4"},
			{Rectype: "TXT", Subname: "_acme", Content: "token"},
			// unsupported type should be silently skipped
			{Rectype: "SRV", Subname: "_sip", Content: "whatever"},
		},
	}

	records, err := resp.IntoLibdnsRecords()
	if err != nil {
		t.Fatal(err)
	}
	// SRV should be skipped, so we expect only 2 records
	if len(records) != 2 {
		t.Errorf("len(records) = %d, want 2", len(records))
	}
}

func TestGetDomainResponseIntoLibdnsRecords_BadSOATTL(t *testing.T) {
	resp := GetDomainResponse{
		SOA:     SOA{TTL: "INVALID"},
		Records: []DNSRecord{{Rectype: "A", Subname: "x", Content: "1.1.1.1"}},
	}
	_, err := resp.IntoLibdnsRecords()
	if err == nil {
		t.Error("expected error from bad SOA TTL")
	}
}
