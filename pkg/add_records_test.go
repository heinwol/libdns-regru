package libdns_regru

import (
	"net/netip"
	"testing"
	"time"

	"github.com/libdns/libdns"
)

// ---------------------------------------------------------------------------
// addRequestFromLibdns
// ---------------------------------------------------------------------------

func TestAddRequestFromLibdns_A(t *testing.T) {
	zone := "example.com"
	rec := libdns.Address{
		Name: "www",
		TTL:  time.Hour,
		IP:   netip.MustParseAddr("1.2.3.4"),
	}
	req, err := addRequestFromLibdns(zone, rec)
	if err != nil {
		t.Fatal(err)
	}
	alias, ok := req.(AddAliasRequest)
	if !ok {
		t.Fatalf("expected AddAliasRequest, got %T", req)
	}
	if alias.IPAddr != "1.2.3.4" {
		t.Errorf("IPAddr = %s, want 1.2.3.4", alias.IPAddr)
	}
	if alias.Subdomain != "www" {
		t.Errorf("Subdomain = %s, want www", alias.Subdomain)
	}
	if len(alias.Domains) != 1 || alias.Domains[0].DName != zone {
		t.Errorf("Domains = %v", alias.Domains)
	}
	if alias.getCommandName() != "add_alias" {
		t.Errorf("command = %s", alias.getCommandName())
	}
}

func TestAddRequestFromLibdns_AAAA(t *testing.T) {
	zone := "example.com"
	rec := libdns.Address{
		Name: "v6host",
		TTL:  time.Hour,
		IP:   netip.MustParseAddr("2001:db8::1"),
	}
	req, err := addRequestFromLibdns(zone, rec)
	if err != nil {
		t.Fatal(err)
	}
	aaaa, ok := req.(AddAAAARequest)
	if !ok {
		t.Fatalf("expected AddAAAARequest, got %T", req)
	}
	if aaaa.IPAddr != "2001:db8::1" {
		t.Errorf("IPAddr = %s", aaaa.IPAddr)
	}
	if aaaa.getCommandName() != "add_aaaa" {
		t.Errorf("command = %s", aaaa.getCommandName())
	}
}

func TestAddRequestFromLibdns_CNAME(t *testing.T) {
	zone := "example.com"
	rec := libdns.CNAME{
		Name:   "alias",
		TTL:    time.Hour,
		Target: "target.example.com.",
	}
	req, err := addRequestFromLibdns(zone, rec)
	if err != nil {
		t.Fatal(err)
	}
	cname, ok := req.(AddCNAMERequest)
	if !ok {
		t.Fatalf("expected AddCNAMERequest, got %T", req)
	}
	if cname.Canonical != "target.example.com." {
		t.Errorf("Canonical = %s", cname.Canonical)
	}
	if cname.getCommandName() != "add_cname" {
		t.Errorf("command = %s", cname.getCommandName())
	}
}

func TestAddRequestFromLibdns_MX(t *testing.T) {
	zone := "example.com"
	prio := uint16(10)
	rec := libdns.MX{
		Name:       "@",
		TTL:        time.Hour,
		Preference: prio,
		Target:     "mail.example.com.",
	}
	req, err := addRequestFromLibdns(zone, rec)
	if err != nil {
		t.Fatal(err)
	}
	mx, ok := req.(AddMXRequest)
	if !ok {
		t.Fatalf("expected AddMXRequest, got %T", req)
	}
	if mx.MailServer != "mail.example.com." {
		t.Errorf("MailServer = %s", mx.MailServer)
	}
	if mx.Priority == nil || *mx.Priority != 10 {
		t.Errorf("Priority = %v", mx.Priority)
	}
	if mx.getCommandName() != "add_mx" {
		t.Errorf("command = %s", mx.getCommandName())
	}
}

func TestAddRequestFromLibdns_NS(t *testing.T) {
	zone := "example.com"
	rec := libdns.NS{
		Name:   "@",
		TTL:    time.Hour,
		Target: "ns1.example.com.",
	}
	req, err := addRequestFromLibdns(zone, rec)
	if err != nil {
		t.Fatal(err)
	}
	ns, ok := req.(AddNSRequest)
	if !ok {
		t.Fatalf("expected AddNSRequest, got %T", req)
	}
	if ns.DNSServer != "ns1.example.com." {
		t.Errorf("DNSServer = %s", ns.DNSServer)
	}
	if ns.getCommandName() != "add_ns" {
		t.Errorf("command = %s", ns.getCommandName())
	}
}

func TestAddRequestFromLibdns_TXT(t *testing.T) {
	zone := "example.com"
	rec := libdns.TXT{
		Name: "_acme-challenge",
		TTL:  time.Hour,
		Text: "some-validation-token",
	}
	req, err := addRequestFromLibdns(zone, rec)
	if err != nil {
		t.Fatal(err)
	}
	txt, ok := req.(AddTXTRequest)
	if !ok {
		t.Fatalf("expected AddTXTRequest, got %T", req)
	}
	if txt.Text != "some-validation-token" {
		t.Errorf("Text = %s", txt.Text)
	}
	if txt.getCommandName() != "add_txt" {
		t.Errorf("command = %s", txt.getCommandName())
	}
}

func TestAddRequestFromLibdns_CAA(t *testing.T) {
	zone := "example.com"
	rec := libdns.CAA{
		Name:  "@",
		TTL:   time.Hour,
		Flags: 0,
		Tag:   "issue",
		Value: "letsencrypt.org",
	}
	req, err := addRequestFromLibdns(zone, rec)
	if err != nil {
		t.Fatal(err)
	}
	caa, ok := req.(AddCAARequest)
	if !ok {
		t.Fatalf("expected AddCAARequest, got %T", req)
	}
	if caa.Tag != "issue" {
		t.Errorf("Tag = %s", caa.Tag)
	}
	if caa.Value != "letsencrypt.org" {
		t.Errorf("Value = %s", caa.Value)
	}
	if caa.getCommandName() != "add_caa" {
		t.Errorf("command = %s", caa.getCommandName())
	}
}

// func TestAddRequestFromLibdns_UnsupportedType(t *testing.T) {
// 	// libdns.RR is a catch-all that won't match any known type switch case
// 	rec := libdns.RR{
// 		Type: "SRV",
// 		Name: "_sip",
// 		Data: "10 20 5060 sipserver.example.com.",
// 	}
// 	_, err := addRequestFromLibdns("example.com", rec)
// 	if err == nil {
// 		t.Error("expected error for unsupported record type SRV")
// 	}
// }

// ---------------------------------------------------------------------------
// getUrl helper
// ---------------------------------------------------------------------------

func TestGetUrl(t *testing.T) {
	cases := []struct {
		req  requestWithName
		want string
	}{
		{AddAliasRequest{}, "/zone/add_alias"},
		{AddAAAARequest{}, "/zone/add_aaaa"},
		{AddCNAMERequest{}, "/zone/add_cname"},
		{AddMXRequest{}, "/zone/add_mx"},
		{AddNSRequest{}, "/zone/add_ns"},
		{AddTXTRequest{}, "/zone/add_txt"},
		{AddCAARequest{}, "/zone/add_caa"},
		{RemoveRecordRequest{}, "/zone/remove_record"},
		{UpdateRecordsRequest{}, "/zone/update_records"},
	}
	for _, tc := range cases {
		got := getUrl(tc.req)
		if got != tc.want {
			t.Errorf("getUrl(%T) = %q, want %q", tc.req, got, tc.want)
		}
	}
}
