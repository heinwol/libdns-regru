package libdns_regru

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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
	Client      resty.Client
	Credentials Credentials
}

func NewRegruClientForTests() (*RegruClient, error) {
	username, found := os.LookupEnv("REGRU_USERNAME")
	if !found {
		log.Fatal("username not found")
	}
	password, found := os.LookupEnv("REGRU_PASSWORD")
	if !found {
		log.Fatal("password not found")
	}
	credentials := Credentials{Username: username, Password: password}

	dialer, err := proxy.SOCKS5("tcp", "localhost:1080", nil, proxy.Direct)
	if err != nil {
		panic(err)
	}
	transport := &http.Transport{
		Dial: dialer.Dial,
	}
	client, err := NewRegruClient(credentials)
	if err != nil {
		return nil, err
	}
	client.Client.SetTransport(transport)
	return client, nil
}

func NewRegruClient(credentials Credentials) (*RegruClient, error) {
	inner := resty.New().
		SetBaseURL("https://api.reg.ru/api/regru2").
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		AddContentTypeEncoder("application/x-www-form-urlencoded", func(w io.Writer, v any) error {
			inputData, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("regru: marshal input_data: %w", err)
			}
			values := url.Values{
				"username":      {credentials.Username},
				"password":      {credentials.Password},
				"input_format":  {"json"},
				"output_format": {"json"},
				"input_data":    {string(inputData)},
			}
			_, err = w.Write([]byte(values.Encode()))
			return err

		}).
		// AddContentTypeDecoder("text/plain", resty.InMemoryJSONUnmarshal).
		// that's actually what reg.ru does instead of returning just "text/plain".
		// Resty matches content-type exactly
		// AddContentTypeDecoder("text/plain; charset=utf-8", resty.InMemoryJSONUnmarshal).
		AddResponseMiddleware(func(c *resty.Client, res *resty.Response) error {
			b := res.Bytes()
			if len(b) == 0 {
				return nil
			}

			var basic APIResponse[any]
			if err := json.Unmarshal(b, &basic); err != nil {
				return fmt.Errorf("reg.ru: unmarshal envelope: %w", err)
			}
			if basic.Result != "success" {
				return fmt.Errorf("reg.ru: %s – %s", basic.ErrorCode, basic.ErrorText)
			}

			if res.Request.Result != nil {
				return json.Unmarshal(b, res.Request.Result)
			}
			return nil
		})
	return &RegruClient{Client: *inner, Credentials: credentials}, nil
}

// Requests:

type DomainRequest struct {
	DName string `json:"dname"`
}

type ZoneGetResourceRecordsRequest struct {
	Domains []DomainRequest `json:"domains"`
}

// Responses:

type APIResponse[T any] struct {
	GeneralResponseErrorInfoAndResult
	Answer       T      `json:"answer,omitempty"`
	CharSet      string `json:"charset,omitempty"`
	MessageStore string `json:"messagestore,omitempty"`
}

type GeneralResponseErrorInfoAndResult struct {
	ErrorCode   string `json:"error_code,omitempty"`
	ErrorText   string `json:"error_text,omitempty"`
	ErrorParams any    `json:"error_params,omitempty"`
	Result      string `json:"result"`
}

type DomainResponse struct {
	GeneralResponseErrorInfoAndResult
	DName     string      `json:"dname"`
	Records   []DNSRecord `json:"rrs"`
	ServiceID json.Number `json:"service_id,omitempty"`
	ServType  string      `json:"servtype,omitempty"`
	SOA       SOA         `json:"soa"`
}

type DNSRecord struct {
	Rectype string      `json:"rectype"`
	Subname string      `json:"subname"`
	Content string      `json:"content"`
	Prio    json.Number `json:"prio,omitempty"`
	State   string      `json:"state,omitempty"`
}

type SOA struct {
	MinimumTTL string `json:"minimum_ttl,omitempty"`
	TTL        string `json:"ttl,omitempty"`
}

type AnswerDomains struct {
	Domains []DomainResponse `json:"domains"`
}

type ZoneGetResourceRecordsResponse = APIResponse[AnswerDomains]

type AddTXTResponse struct {
	GeneralResponseErrorInfoAndResult
	Answer AnswerDomains `json:"answer"`
}

type GetRecordsResponse struct {
	GeneralResponseErrorInfoAndResult
	Answer AnswerDomains `json:"answer"`
}
