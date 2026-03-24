package main

import (
	"encoding/json"
	"fmt"

	r "github.com/heinwol/libdns-regru/pkg"
)

func some_test() error {
	client, _ := r.NewRegruClientForTests()

	var respBody r.ZoneGetResourceRecordsResponse
	res, err := client.Client.R().SetBody(r.ZoneGetResourceRecordsRequest{
		Domains: []r.DomainRequest{{
			DName: "helix-info.space",
		}},
	}).
		SetResult(&respBody).
		Post("/zone/get_resource_records")
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("%s", err)
	}

	r.PrettyPrint(respBody)

	var _ = res
	return nil
}

func some_test2() error {
	var raw = `{
		"answer": {
			"domains": [
			{
				"dname": "helix-info.space",
				"result": "success",
				"rrs": [
					{
						"content": "2.56.178.233",
						"prio": 0,
						"rectype": "TXT",
						"state": "A",
						"subname": "@"
					}
				],
				"soa": {
					"minimum_ttl": "3h",
					"ttl": "1m"
				}
			}
			]
		},
		"charset": "utf-8",
		"result": "success"
		}`
	var resp r.ZoneGetResourceRecordsResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		fmt.Printf("unmarshal error: %v\n", err)
		return err
	}
	r.PrettyPrint("result: %s\n", resp.Result)
	r.PrettyPrint("domains: %+v\n", resp.Answer.Domains)
	r.PrettyPrint("first record: %+v\n", resp.Answer.Domains[0].Records)
	r.PrettyPrint(resp.Answer.Domains[0].IntoLibnsRecords())
	return nil
}

func main() {
	some_test2()
}
