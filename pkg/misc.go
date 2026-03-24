package libdns_regru

import (
	"encoding/json"
	"fmt"
)

func MustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func PrettyPrint(v ...any) {
	for _, v := range v {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		fmt.Println(string(b))
	}
}
