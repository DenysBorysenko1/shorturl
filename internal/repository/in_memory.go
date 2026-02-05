package repository

import (
	"sync"
)

var (
	data  = make(map[string]string)
	mu     sync.Mutex
)

func Save(url T) {
	// mu.Lock()
	// todo: hash generate
	// links[hash] = url
	// mu.Unlock()
}

func Get(id string) {
	// mu.Lock()
	// url, ok := links[id]
	// mu.Unlock()

}
