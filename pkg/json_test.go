package libdns_regru

import (
	"encoding/json"
	"testing"
)

// ---------------------------------------------------------------------------
// DNSRecord custom JSON marshal/unmarshal
// ---------------------------------------------------------------------------

func TestDNSRecordUnmarshalJSON_PrioField(t *testing.T) {
	// Test server response that uses "prio" (not "priority")
	raw := `{"rectype":"MX","subname":"@","content":"mail.example.com.","prio":"10","state":"A"}`
	var rec DNSRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if rec.Priority == nil {
		t.Fatal("Priority is nil")
	}
	if *rec.Priority != 10 {
		t.Errorf("Priority = %d, want 10", *rec.Priority)
	}
}

func TestDNSRecordUnmarshalJSON_PriorityField(t *testing.T) {
	// Some responses use "priority" key
	raw := `{"rectype":"MX","subname":"@","content":"mail.example.com.","priority":"20"}`
	var rec DNSRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if rec.Priority == nil {
		t.Fatal("Priority is nil")
	}
	if *rec.Priority != 20 {
		t.Errorf("Priority = %d, want 20", *rec.Priority)
	}
}

func TestDNSRecordUnmarshalJSON_NoPriority(t *testing.T) {
	raw := `{"rectype":"A","subname":"www","content":"1.2.3.4"}`
	var rec DNSRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if rec.Priority != nil {
		t.Errorf("expected nil Priority, got %d", *rec.Priority)
	}
	if rec.Rectype != "A" {
		t.Errorf("Rectype = %s, want A", rec.Rectype)
	}
}

func TestDNSRecordMarshalJSON_WithPriority(t *testing.T) {
	prio := uint16(5)
	rec := DNSRecord{
		Rectype:  "MX",
		Subname:  "@",
		Content:  "mail.example.com.",
		Priority: &prio,
	}
	b, err := json.Marshal(rec)
	if err != nil {
		t.Fatal(err)
	}
	// The output should use "priority" key (not "prio")
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["priority"]; !ok {
		t.Errorf("expected 'priority' key in marshaled output, got: %s", b)
	}
	// Should NOT have a separate "prio" key
	if _, ok := m["prio"]; ok {
		t.Errorf("unexpected 'prio' key in marshaled output")
	}
}

// ---------------------------------------------------------------------------
// APIResponse / GeneralResponseErrorInfoAndResult
// ---------------------------------------------------------------------------

func TestAPIResponseIntoError_Success(t *testing.T) {
	info := GeneralResponseErrorInfoAndResult{Result: "success"}
	if err := info.intoError(); err != nil {
		t.Errorf("expected nil error for result=success, got: %v", err)
	}
}

func TestAPIResponseIntoError_Error(t *testing.T) {
	info := GeneralResponseErrorInfoAndResult{
		Result:    "error",
		ErrorCode: "SOME_ERROR",
		ErrorText: "some error text",
	}
	err := info.intoError()
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.ErrorCode != "SOME_ERROR" {
		t.Errorf("ErrorCode = %s", err.ErrorCode)
	}
}

func TestAPIResponseUnmarshal(t *testing.T) {
	raw := `{
		"result": "success",
		"charset": "utf-8",
		"answer": {
			"domains": [
				{
					"dname": "example.com",
					"result": "success",
					"rrs": [
						{"rectype": "A", "subname": "www", "content": "1.2.3.4", "prio": "0", "state": "A"}
					],
					"soa": {"ttl": "1d", "minimum_ttl": "12h"}
				}
			]
		}
	}`

	var resp GetResourceRecordsResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Result != "success" {
		t.Errorf("Result = %s", resp.Result)
	}
	if len(resp.Answer.Domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(resp.Answer.Domains))
	}
	dom := resp.Answer.Domains[0]
	if dom.DName != "example.com" {
		t.Errorf("DName = %s", dom.DName)
	}
	if dom.SOA.TTL != "1d" {
		t.Errorf("SOA.TTL = %s", dom.SOA.TTL)
	}
	if len(dom.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(dom.Records))
	}
	if dom.Records[0].Content != "1.2.3.4" {
		t.Errorf("Content = %s", dom.Records[0].Content)
	}
}
