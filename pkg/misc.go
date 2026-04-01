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
	// f          func() (T, error)
}

// func OnceCellWithFunc[T any](f func() (T, error)) OnceCell[T] {
// 	return OnceCell[T]{f: f}
// }

func (self *onceCell[T]) Do(f func() (*T, error)) (*T, error) {
	// if self.f == nil {
	// 	return nil, fmt.Errorf("internal function is absent")
	// }
	self.once.Do(func() {
		self.Inner, self.init_error = f()
		// self.inner, self.init_error = &inner, err
	})
	return self.Inner, self.init_error
}

// I know it's not idiomatic; I don't care
type mutexWrapper[T any] struct {
	inner T
	mutex sync.Mutex
}

func withLock[T, F any](m *mutexWrapper[T], f func(*T) (F, error)) (F, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return f(&m.inner)
}
