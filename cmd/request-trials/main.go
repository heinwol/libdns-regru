package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	r "github.com/heinwol/libdns-regru/pkg"
)

func some_test() error {
	client, _ := r.NewRegruClientForTests()

	var respBody r.GetResourceRecordsResponse
	res, err := client.Client.R().SetBody(r.GetResourceRecordsRequest{
		Domains: []r.GetDomainRequest{{
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
	records, err := os.ReadFile("data/current_records.json")
	if err != nil {
		log.Fatal(err)
	}
	var resp r.GetResourceRecordsResponse
	if err := json.Unmarshal(records, &resp); err != nil {
		fmt.Printf("unmarshal error: %v\n", err)
		return err
	}
	r.PrettyPrint(resp.Answer.Domains[0].IntoLibnsRecords())
	return nil
}

func main() {
	some_test2()
}
