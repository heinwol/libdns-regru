package libdns_regru

import (
	"encoding/json"
	"fmt"
	"sync"
)

func MustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func PrettyJsonBytes(b []byte) (string, error) {
	var obj any
	err := json.Unmarshal(b, &obj)
	prettyJSON, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return "", nil
	}
	return string(prettyJSON), nil
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

type onceCell[T any] struct {
	Inner      *T
	once       sync.Once
	init_error error
}

func (self *onceCell[T]) Do(f func() (*T, error)) (*T, error) {
	self.once.Do(func() {
		self.Inner, self.init_error = f()
	})
	return self.Inner, self.init_error
}
