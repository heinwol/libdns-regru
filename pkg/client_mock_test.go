package libdns_regru

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/libdns/libdns"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newTestClient builds a RegruClient pointing at the given httptest server URL.
func newTestClient(t *testing.T, serverURL string) *RegruClient {
	t.Helper()
	creds := Credentials{Username: "test", Password: "test"}
	client, err := NewRegruClient(creds)
	if err != nil {
		t.Fatalf("NewRegruClient: %v", err)
	}
	client.Client.SetBaseURL(serverURL)
	return client
}

// parseInputData decodes the form-encoded body and returns the parsed
// input_data JSON object as a map.
func parseInputData(t *testing.T, body []byte) map[string]any {
	t.Helper()
	values, err := url.ParseQuery(string(body))
	if err != nil {
		t.Fatalf("parseInputData: ParseQuery: %v", err)
	}
	raw := values.Get("input_data")
	if raw == "" {
		t.Fatal("parseInputData: input_data field is empty")
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("parseInputData: unmarshal: %v", err)
	}
	return m
}

// jsonResponse writes a JSON body.
func jsonResponse(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

// Small record-building helpers.
func libdnsTXT(name, text string) libdns.TXT {
	return libdns.TXT{Name: name, TTL: time.Hour, Text: text}
}

func libdnsAddress4(name, ip string) libdns.Address {
	return libdns.Address{Name: name, TTL: time.Hour, IP: netip.MustParseAddr(ip)}
}

// ---------------------------------------------------------------------------
// Client middleware: credential injection
// ---------------------------------------------------------------------------

func TestClientInjectsCredentials(t *testing.T) {
	var capturedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		jsonResponse(w, map[string]any{"result": "success", "answer": map[string]any{}})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	var result any
	_, _ = client.Client.R().SetBody(map[string]any{"domains": []any{}}).SetResult(&result).Post("/zone/nop")

	values, _ := url.ParseQuery(string(capturedBody))
	if values.Get("username") != "test" {
		t.Errorf("username = %q, want test", values.Get("username"))
	}
	if values.Get("password") != "test" {
		t.Errorf("password = %q, want test", values.Get("password"))
	}
	if values.Get("input_format") != "json" {
		t.Errorf("input_format = %q, want json", values.Get("input_format"))
	}
	if values.Get("output_format") != "json" {
		t.Errorf("output_format = %q, want json", values.Get("output_format"))
	}
}

func TestClientMiddleware_ReturnsErrorOnFailResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]any{
			"result":     "error",
			"error_code": "ACCESS_DENIED",
			"error_text": "access denied",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.Client.R().SetBody(map[string]any{}).Post("/zone/nop")
	if err == nil {
		t.Fatal("expected error from middleware, got nil")
	}
	apiErr, ok := err.(*APIResponseError)
	if !ok {
		t.Fatalf("expected *APIResponseError, got %T: %v", err, err)
	}
	if apiErr.ErrorCode != "ACCESS_DENIED" {
		t.Errorf("ErrorCode = %q, want ACCESS_DENIED", apiErr.ErrorCode)
	}
}

// ---------------------------------------------------------------------------
// GetZoneRecords
// ---------------------------------------------------------------------------

func TestGetZoneRecords_Success(t *testing.T) {
	zone := "test.ru"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/zone/get_resource_records") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, map[string]any{
			"result":  "success",
			"charset": "utf-8",
			"answer": map[string]any{
				"domains": []any{
					map[string]any{
						"dname":  zone,
						"result": "success",
						"rrs": []any{
							map[string]any{"rectype": "A", "subname": "www", "content": "1.2.3.4", "prio": "0", "state": "A"},
							map[string]any{"rectype": "TXT", "subname": "_acme", "content": "token", "prio": "0", "state": "A"},
						},
						"soa": map[string]any{"ttl": "1d", "minimum_ttl": "12h"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	resp, err := client.GetZoneRecords(t.Context(), zone)
	if assert.NoError(t, err) && assert.Equal(t, resp.Result, "success") {
		assert.Equal(t, resp.DName, testZone)
		records := resp.Records
		if assert.Len(t, records, 2) {
			assert.Equal(t, records[0].Rectype, "A")
			assert.Equal(t, records[1].Rectype, "TXT")
		}
		assert.Equal(t, resp.SOA.TTL, "1d")
	}
}

func TestGetZoneRecords_EmptyDomains(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]any{
			"result":  "success",
			"charset": "utf-8",
			"answer":  map[string]any{"domains": []any{}},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.GetZoneRecords(t.Context(), "nonexistent.com")
	if err == nil {
		t.Error("expected error for empty domains list")
	}
}

func TestGetZoneRecords_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]any{
			"result":     "error",
			"error_code": "DOMAIN_NOT_FOUND",
			"error_text": "domain not found",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.GetZoneRecords(t.Context(), "notfound.com")
	if err == nil {
		t.Error("expected error from API error response")
	}
}

// ---------------------------------------------------------------------------
// AddZoneRecord
// ---------------------------------------------------------------------------

func TestAddZoneRecord_TXT(t *testing.T) {
	var capturedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		if !strings.HasSuffix(r.URL.Path, "/zone/add_txt") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, map[string]any{
			"result":  "success",
			"charset": "utf-8",
			"answer": map[string]any{
				"domains": []any{
					map[string]any{"dname": "test.ru", "result": "success"},
				},
			},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	rec, err := client.AddZoneRecord(t.Context(), "test.ru", libdnsTXT("_acme-challenge", "mytoken"))
	if err != nil {
		t.Fatal(err)
	}
	if rec.Result != "success" {
		t.Errorf("Result = %s", rec.Result)
	}

	data := parseInputData(t, capturedBody)
	if data["text"] != "mytoken" {
		t.Errorf("text field = %v", data["text"])
	}
	if data["subdomain"] != "_acme-challenge" {
		t.Errorf("subdomain = %v", data["subdomain"])
	}
}

func TestAddZoneRecord_A_UsesAddAlias(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		jsonResponse(w, map[string]any{
			"result": "success",
			"answer": map[string]any{
				"domains": []any{
					map[string]any{"dname": "test.ru", "result": "success"},
				},
			},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.AddZoneRecord(t.Context(), "test.ru", libdnsAddress4("myhost", "5.6.7.8"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(capturedPath, "/zone/add_alias") {
		t.Errorf("expected add_alias path, got %s", capturedPath)
	}
}

// ---------------------------------------------------------------------------
// RemoveZoneRecord
// ---------------------------------------------------------------------------

func TestRemoveZoneRecord(t *testing.T) {
	var capturedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		if !strings.HasSuffix(r.URL.Path, "/zone/remove_record") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, map[string]any{
			"result": "success",
			"answer": map[string]any{
				"domains": []any{
					map[string]any{"dname": "test.ru", "result": "success"},
				},
			},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.RemoveZoneRecord(t.Context(), "test.ru", libdnsTXT("_acme-challenge", "token"))
	if err != nil {
		t.Fatal(err)
	}

	data := parseInputData(t, capturedBody)
	if data["subdomain"] != "_acme-challenge" {
		t.Errorf("subdomain = %v", data["subdomain"])
	}
	if data["record_type"] != "TXT" {
		t.Errorf("record_type = %v", data["record_type"])
	}
}

// ---------------------------------------------------------------------------
// UpdateSOA
// ---------------------------------------------------------------------------

func TestUpdateSOA(t *testing.T) {
	var capturedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		if !strings.HasSuffix(r.URL.Path, "/zone/update_soa") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, map[string]any{
			"result": "success",
			"answer": map[string]any{
				"domains": []any{
					map[string]any{"dname": "test.ru", "result": "success"},
				},
			},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.UpdateSOA(t.Context(), "test.ru", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	data := parseInputData(t, capturedBody)
	if data["ttl"] != "3600" {
		t.Errorf("ttl = %v, want 3600", data["ttl"])
	}
	domains, _ := data["domains"].([]any)
	if len(domains) != 1 {
		t.Fatalf("expected 1 domain in request, got %v", domains)
	}
	dom := domains[0].(map[string]any)
	if dom["dname"] != "test.ru" {
		t.Errorf("dname = %v", dom["dname"])
	}
}

// ---------------------------------------------------------------------------
// UpdateZoneRecords
// ---------------------------------------------------------------------------

func TestUpdateZoneRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/zone/update_records") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, map[string]any{
			"result": "success",
			"answer": map[string]any{
				"domains": []any{
					map[string]any{
						"dname":  "test.ru",
						"result": "success",
						"action_list": []any{
							map[string]any{"action": "add_txt", "result": "success"},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	resp, err := client.UpdateZoneRecords(t.Context(), "test.ru", []libdns.Record{
		libdnsTXT("_acme-challenge", "token"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result != "success" {
		t.Errorf("Result = %s", resp.Result)
	}
}

// ---------------------------------------------------------------------------
// AnalyzeUpdateResponse
// ---------------------------------------------------------------------------

func makeUpdateResponse(zone string, actionResults ...map[string]any) *UpdateResponse {
	actions := make([]UpdateActionResponse, 0, len(actionResults))
	for _, a := range actionResults {
		result, _ := a["result"].(string)
		errCode, _ := a["error_code"].(string)
		name, _ := a["action"].(string)
		actions = append(actions, UpdateActionResponse{
			GeneralResponseErrorInfoAndResult: GeneralResponseErrorInfoAndResult{
				Result:    result,
				ErrorCode: errCode,
			},
			ActionName: name,
		})
	}
	return &UpdateResponse{
		GeneralResponseErrorInfoAndResult: GeneralResponseErrorInfoAndResult{Result: "success"},
		Answer: UpdateDomainsAnswer{
			Domains: []UpdateZoneResponse{
				{
					GeneralResponseErrorInfoAndResult: GeneralResponseErrorInfoAndResult{Result: "success"},
					DName:                             zone,
					ActionList:                        actions,
				},
			},
		},
	}
}

func TestAnalyzeUpdateResponse_AllSuccess(t *testing.T) {
	zone := "test.ru"
	records := []libdns.Record{libdnsTXT("_acme", "token")}
	resp := makeUpdateResponse(zone,
		map[string]any{"action": "add_txt", "result": "success"},
	)

	updated, err := AnalyzeUpdateResponse(&resp.Answer.Domains[0], zone, records)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(updated) != 1 {
		t.Errorf("len(updated) = %d, want 1", len(updated))
	}
}

func TestAnalyzeUpdateResponse_ActionError(t *testing.T) {
	zone := "test.ru"
	records := []libdns.Record{libdnsTXT("_acme", "token")}
	resp := makeUpdateResponse(zone,
		map[string]any{"action": "add_txt", "result": "error", "error_code": "SUBD_INVALID"},
	)
	_, err := AnalyzeUpdateResponse(&resp.Answer.Domains[0], zone, records)
	if err == nil {
		t.Error("expected error from action-level failure")
	}
}

func TestAnalyzeUpdateResponse_RecordNotInActionList(t *testing.T) {
	zone := "test.ru"
	records := []libdns.Record{libdnsTXT("_acme", "token")}
	// action list has no "add_txt" entry
	resp := makeUpdateResponse(zone,
		map[string]any{"action": "add_alias", "result": "success"},
	)
	updated, err := AnalyzeUpdateResponse(&resp.Answer.Domains[0], zone, records)
	// Should produce an error (record not found in answer) and no updated records
	if err == nil {
		t.Error("expected error because TXT was not in action_list")
	}
	if len(updated) != 0 {
		t.Errorf("expected 0 updated records, got %d", len(updated))
	}
}
