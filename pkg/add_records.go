package libdns_regru

// Requests

type GeneralAddDomainRequest struct {
	Domains   []GeneralZoneRequest `json:"domains"`
	Subdomain string               `json:"subdomain"`
}

//

type AddAliasRequest struct {
	GeneralAddDomainRequest
	IPAddr string `json:"ipaddr"`
}

type AddAAAARequest struct {
	GeneralAddDomainRequest
	IPAddr string `json:"ipaddr"`
}

type AddCAARequest struct {
	GeneralAddDomainRequest
	Flags uint8  `json:"flags"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

type AddCNAMERequest struct {
	GeneralAddDomainRequest
	Canonical string `json:"canonical_name"`
}

type AddMXRequest struct {
	GeneralAddDomainRequest
	MailServer string `json:"mail_server"`
	Priority   *uint8 `json:"priority,omitempty"`
}

type AddNSRequest struct {
	GeneralAddDomainRequest
	DNSServer    string `json:"dns_server"`
	RecordNumber *int   `json:"record_number,omitempty"`
}

type AddTXTRequest struct {
	GeneralAddDomainRequest
	Text string `json:"text"`
}

// Responses:

type AddResponse = APIResponse[AddDomainsAnswer]

type AddDomainsAnswer struct {
	Domains []AddDomainResponse `json:"domains"`
}

type AddDomainResponse struct {
	GeneralResponseErrorInfoAndResult
	DName     string `json:"dname"`
	ServiceID string `json:"service_id,omitempty"`
}

// func (self *RegruClient) AddZoneRecord(ctx context.Context, zone string, record libdns.Record) (*AddDomainResponse, error) {
// 	// libdns.NS
// }
