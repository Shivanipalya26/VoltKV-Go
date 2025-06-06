package store

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Store struct {
	data     map[string]string
	hashes   map[string]map[string]string
	lists    map[string][]string
	waiters  map[string][]chan [2]string
	expiries map[string]time.Time
	mu       sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		data:     make(map[string]string),
		hashes:   make(map[string]map[string]string),
		lists:    make(map[string][]string),
		waiters:  make(map[string][]chan [2]string),
		expiries: make(map[string]time.Time),
	}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	exp, hasExpiry := s.expiries[key]
	isExpired := hasExpiry && time.Now().After(exp)
	s.mu.RUnlock()

	if isExpired {
		s.mu.Lock()
		delete(s.data, key)
		delete(s.expiries, key)
		s.mu.Unlock()
		return "", false
	}

	s.mu.RLock()
	val, ok := s.data[key]
	s.mu.RUnlock()
	return val, ok
}

func (s *Store) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; exists {
		delete(s.data, key)
		return true
	}
	return false
}

func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.data[key]
	return exists
}

func (s *Store) Expire(key string, seconds int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; !exists {
		return false
	}
	s.expiries[key] = time.Now().Add(time.Duration(seconds) * time.Second)
	return true
}

func (s *Store) HSet(key string, fields map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hashes[key]; !exists {
		s.hashes[key] = make(map[string]string)
	}
	for field, value := range fields {
		s.hashes[key][field] = value
	}

}

func (s *Store) HGet(key, field string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if hash, exists := s.hashes[key]; exists {
		val, ok := hash[field]
		return val, ok
	}
	return "", false
}

func (s *Store) HGetAll(key string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if hash, exists := s.hashes[key]; exists {
		copy := make(map[string]string)
		for k, v := range hash {
			copy[k] = v
		}
		return copy
	}
	return nil
}

func (s *Store) LPush(key string, values ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.lists[key]
	list = append(reverse(values), list...)
	s.lists[key] = list

	for len(list) > 0 && len(s.waiters[key]) > 0 {
		waitCh := s.waiters[key][0]
		s.waiters[key] = s.waiters[key][1:]

		val := list[0]
		list = list[1:]
		waitCh <- [2]string{key, val}
	}

	log.Printf("LPUSH array %v", list)

	return len(list)
}

func (s *Store) RPush(key string, values ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.lists[key]
	list = append(list, values...)
	s.lists[key] = list

	log.Printf("RPUSH array %v", list)

	return len((list))

}

func (s *Store) LPop(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.lists[key]
	if len(list) == 0 {
		return "", false
	}
	val := list[0]
	s.lists[key] = list[1:]
	return val, true
}

func (s *Store) RPop(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.lists[key]
	if len(list) == 0 {
		return "", false
	}
	lastInd := len(list) - 1
	val := list[lastInd]
	s.lists[key] = list[:lastInd]
	return val, true
}

func reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return s
}

func (s *Store) RegisterWaiter(key string, ch chan [2]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.waiters[key] = append(s.waiters[key], ch)
}

func (s *Store) StartCleaner(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			s.mu.Lock()
			now := time.Now()
			for key, exp := range s.expiries {
				if now.After(exp) {
					delete(s.data, key)
					delete(s.expiries, key)
					fmt.Println("[cleaner] Deleted expired key:", key)
				}
			}
			s.mu.Unlock()
		}
	}()
}
