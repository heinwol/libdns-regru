package main

import (
	"fmt"

	r "github.com/heinwol/libdns-regru/pkg"
)

func some_test() error {

	client, _ := r.NewRegruClient()

	res, err := client.WithPayload(r.ZoneGetResourceRecordsRequest{
		Domains: []r.DomainRequest{{
			DName: "helix-info.space",
		}},
	}).
		Post("/zone/get_resource_records")
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(res)
	return nil
}

func main() {
	some_test()
}
