// Integration tests that hit the real reg.ru API using the public
// test account (username=test, password=test).
//
// The test account returns synthetic but structurally valid responses without
// applying real changes, so these tests are safe to run at any time.
//
// Run with:
//
//	go test ./pkg/... -run Integration -v
//
// By default, tests require the environment variable REGRU_INTEGRATION=1 to
// avoid accidentally running them in CI without network access.

package libdns_regru

import (
	"context"
	"testing"
	"time"

	"github.com/libdns/libdns"
	"github.com/stretchr/testify/assert"
)

// testZone is the domain known to be available under the reg.ru test account.
const testZone string = "test.ru"

// integrationClient returns a real RegruClient using test credentials, or
// skips the test if REGRU_INTEGRATION is not set.
func integrationClient(t *testing.T) *RegruClient {
	t.Helper()
	// TODO: revert this when done
	// if os.Getenv("REGRU_INTEGRATION") == "" {
	// 	t.Skip("set REGRU_INTEGRATION=1 to run integration tests")
	// }
	client, err := NewRegruClient(Credentials{Username: "test", Password: "test"})
	if err != nil {
		t.Fatalf("NewRegruClient: %v", err)
	}
	return client
}

// ---------------------------------------------------------------------------
// zone/get_resource_records
// ---------------------------------------------------------------------------

func TestIntegration_GetZoneRecords(t *testing.T) {
	client := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := client.GetZoneRecords(ctx, testZone)
	if assert.NoError(t, err) && assert.Equal(t, resp.Result, "success") {
		assert.Equal(t, resp.DName, testZone)
		assert.Equal(t, resp.DName, testZone)

		if assert.NotEmpty(t, resp) {
			_, err = fromRegruTTL(resp.SOA.TTL)
			assert.Nil(t, err)
			_, err = fromRegruTTL(resp.SOA.MinimumTTL)
			assert.Nil(t, err)
			assert.NotEmpty(t, resp.Records[0].Subname)
		}

	}
}

func TestIntegration_GetZoneRecords_IntoLibdns(t *testing.T) {
	client := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := client.GetZoneRecords(ctx, testZone)
	if assert.NoError(t, err) {
		records, err := resp.IntoLibdnsRecords()
		if assert.NoError(t, err) {
			if len(records) == 0 {
				t.Log("warning: no records returned by test zone (unexpected but not fatal)")
			} else {
				for _, r := range records {
					rr := r.RR()
					assert.NotEmpty(t, rr.Type)
					assert.NotEmpty(t, rr.Name)
					assert.NotEmpty(t, rr.TTL)
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// zone/add_txt  (test mode — no real change applied)
// ---------------------------------------------------------------------------

func TestIntegration_AddTXTRecord(t *testing.T) {
	client := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rec := libdnsTXT("_acme-challenge-integration-test", "integration-test-token")
	resp, err := client.AddZoneRecord(ctx, testZone, rec)
	if assert.NoError(t, err) && assert.Equal(t, resp.Result, "success") {
		assert.Equal(t, resp.DName, testZone)
		assert.Equal(t, resp.DName, testZone)
	}
}

// ---------------------------------------------------------------------------
// zone/add_alias  (IPv4, test mode)
// ---------------------------------------------------------------------------

func TestIntegration_AddARecord(t *testing.T) {
	client := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rec := libdnsAddress4("integration-test-host", "1.2.3.4")
	resp, err := client.AddZoneRecord(ctx, testZone, rec)
	if assert.NoError(t, err) && assert.Equal(t, resp.Result, "success") {
		assert.Equal(t, resp.DName, testZone)
		assert.Equal(t, resp.DName, testZone)
	}
}

// ---------------------------------------------------------------------------
// zone/remove_record  (test mode)
// ---------------------------------------------------------------------------

func TestIntegration_RemoveTXTRecord(t *testing.T) {
	client := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rec := libdns.TXT{Name: "_acme-challenge-integration-test"}
	resp, err := client.RemoveZoneRecord(ctx, testZone, rec)
	if assert.NoError(t, err) && assert.Equal(t, resp.Result, "success") {
		assert.Equal(t, resp.DName, testZone)
	}
}

// ---------------------------------------------------------------------------
// zone/update_soa  (test mode)
// ---------------------------------------------------------------------------

// TestIntegration_UpdateSOA exposes a known API inconsistency: the reg.ru test
// account returns service_id as a JSON number, but UpdateSOADomainResponse
// (and UpdateZoneResponse) declare it as a string. This causes an unmarshal
// error that the production account does not trigger (it returns a string).
// Track this in the source: UpdateSOADomainResponse.ServiceID should use
// json.Number or *json.Number to handle both.
func TestIntegration_UpdateSOA(t *testing.T) {
	client := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := client.UpdateSOA(ctx, testZone, time.Hour)
	if assert.NoError(t, err) && assert.Equal(t, resp.Result, "success") {
		assert.Equal(t, resp.DName, testZone)
		assert.Equal(t, resp.DName, testZone)
	}
}

// ---------------------------------------------------------------------------
// zone/update_records  (test mode)
// ---------------------------------------------------------------------------

// Same service_id type mismatch as TestIntegration_UpdateSOA.
func TestIntegration_UpdateZoneRecords(t *testing.T) {
	client := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	records := []libdns.Record{
		libdnsTXT("_integration-update", "value1"),
		libdns.MX{Name: "main", TTL: time.Hour, Target: "my-mail@test.com"},
	}
	resp, err := client.UpdateZoneRecords(ctx, testZone, records)
	if assert.NoError(t, err) && assert.Equal(t, resp.Result, "success") {
		assert.Equal(t, resp.DName, testZone)
		updated, err := AnalyzeUpdateResponse(resp, testZone, records)
		if assert.NoError(t, err) &&
			assert.Len(t, updated, 2) {
			assert.Equal(t, updated[0].RR().Type, "TXT")
			assert.Equal(t, updated[1].RR().Type, "MX")
		}

	}

}
