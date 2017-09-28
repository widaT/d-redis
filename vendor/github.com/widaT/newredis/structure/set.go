package structure

import (
	"sync"
	"math/rand"
	"time"
)

type Set struct {
	mu sync.Mutex
	Key   string
	Mset   map[string]struct{}
}

func (s *Set) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Mset)
}

func (s *Set) Add(key string) int  {
	s.mu.Lock()
	defer  s.mu.Unlock()

	if _,found := s.Mset[key];!found  {
		s.Mset[key] = struct{}{}
	}else{
		return 0
	}
	return 1
}

func (s *Set)Del(key string) error  {
	s.mu.Lock()
	defer  s.mu.Unlock()
	delete(s.Mset,key)
	return nil
}

func (s *Set)Members() *[][]byte {
	s.mu.Lock()
	defer  s.mu.Unlock()
	var ret [][]byte

	for key,_:=range s.Mset {
		ret = append(ret,[]byte(key))
	}
	return &ret
}

func (s *Set) Exists(key string) int {
	s.mu.Lock()
	defer  s.mu.Unlock()

	if _,found:= s.Mset[key] ; found {
		return 1
	}
	return 0
}

func (s *Set) RandomKey()  string  {
	var keys []string
	for k,_ :=range s.Mset {
		keys = append(keys,k)
	}
	len := len(keys)
	if len == 0 {
		return ""
	}

	if len == 1 {
		return keys[0]
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	index :=  r.Intn(len-1)
	return keys[index]
}

func NewSset(key string) *Set {
	return &Set{
		Key:   key,
		Mset: make(map[string]struct{}),
	}
}