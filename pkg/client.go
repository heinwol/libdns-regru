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

			// // TODO: just for debug purposes, unsafe for prod
			// if _, exists := os.LookupEnv("LOG_VERBOSE"); exists {
			// 	b_fmt, err := PrettyJsonBytes(b)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	slog.Warn(string(b_fmt))
			// }

			var api_response APIResponse[any]
			if err := json.Unmarshal(b, &api_response); err != nil {
				return fmt.Errorf("reg.ru: unmarshal envelope: %w", err)
			}
			if it := api_response.intoError(); it != nil {
				return it
			}

			// TODO: is this some bs or does it really matter?
			if res.Request.Result != nil {
				return json.Unmarshal(b, res.Request.Result)
			}
			return nil
		})
	return &RegruClient{Client: *inner, Credentials: credentials}, nil
}
