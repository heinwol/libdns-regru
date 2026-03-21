package libdns_regru

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/net/proxy"
	"resty.dev/v3"
)

// Client:

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegruClient struct {
	Inner       resty.Client
	Credentials Credentials
}

func NewRegruClient() (*RegruClient, error) {
	username, found := os.LookupEnv("REGRU_USERNAME")
	if !found {
		log.Fatal("username not found")
	}
	password, found := os.LookupEnv("REGRU_PASSWORD")
	if !found {
		log.Fatal("password not found")
	}
	cred := Credentials{Username: username, Password: password}

	dialer, err := proxy.SOCKS5("tcp", "localhost:1080", nil, proxy.Direct)
	if err != nil {
		panic(err)
	}
	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	inner := resty.New().
		SetBaseURL("https://api.reg.ru/api/regru2").
		SetTransport(transport)
	return &RegruClient{Inner: *inner, Credentials: cred}, nil
}

func (self *RegruClient) WithPayload(payload any) *resty.Request {
	marshaled_payload := MustJSON(payload)
	return self.Inner.R().SetFormData(map[string]string{
		"username":     self.Credentials.Username,
		"password":     self.Credentials.Password,
		"input_format": "json",
		"input_data":   marshaled_payload,
	})
}

// Requests:

type DomainRequest struct {
	DName string `json:"dname"`
}

type ZoneGetResourceRecordsRequest struct {
	Domains []DomainRequest `json:"domains"`
}

// Responses:

type DNSRecord struct {
	RecordID   int64  `json:"record_id"`
	RecordType string `json:"record_type"`
	Subname    string `json:"subname"`
	Content    string `json:"content"`
	TTL        int    `json:"ttl"`
	Priority   int    `json:"priority,omitempty"`
}

type DomainResponse struct {
	DName   string      `json:"dname"`
	Result  string      `json:"result"`
	Records []DNSRecord `json:"rrs,omitempty"`
}

type AnswerDomains struct {
	Domains []DomainResponse `json:"domains"`
}

type ErrorItem struct {
	ErrorCode string `json:"error_code"`
	ErrorText string `json:"error_text"`
}

type ZoneGetResourceRecordsResponse struct {
	Result    string        `json:"result"`
	Answer    AnswerDomains `json:"answer"`
	ErrorCode string        `json:"error_code,omitempty"`
	ErrorText string        `json:"error_text,omitempty"`
}

type ResponseStatus struct {
	Result string `json:"result"`
}

type BasicResponse struct {
	Result      string      `json:"result"`
	Answer      any         `json:"answer,omitempty"`
	ErrorCode   string      `json:"error_code,omitempty"`
	ErrorText   string      `json:"error_text,omitempty"`
	ErrorParams []ErrorItem `json:"error_params,omitempty"`
}

type AddTXTResponse struct {
	Result string        `json:"result"`
	Answer AnswerDomains `json:"answer"`
}

type GetRecordsResponse struct {
	Result string        `json:"result"`
	Answer AnswerDomains `json:"answer"`
}
