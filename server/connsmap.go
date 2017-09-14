package server

import (
	"sync"
)

type ConnMap struct {
	sync.Mutex
	v map[string]Conn
}

func (s *ConnMap) Len() int {
	s.Lock()
	defer s.Unlock()
	return len(s.v)
}

func (s *ConnMap) Add(key string,c Conn) int  {
	s.Lock()
	defer  s.Unlock()
	if _,found := s.v[key];!found  {
		s.v[key] = c
	}else{
		return 0
	}
	return 1
}

func (s *ConnMap) Get(key string)  Conn  {
	s.Lock()
	defer  s.Unlock()
	if _,found := s.v[key];!found  {
		return nil
	}
	return s.v[key]
}


func (s *ConnMap)Del(key string) error  {
	s.Lock()
	defer  s.Unlock()
	delete(s.v,key)
	return nil
}

func (s *ConnMap)Members() *[][]byte {
	s.Lock()
	defer  s.Unlock()
	var ret [][]byte

	for key,_:=range s.v {
		ret = append(ret,[]byte(key))
	}
	return &ret
}

func (s *ConnMap) Exists(key string) bool {
	s.Lock()
	defer  s.Unlock()
	if _,found:= s.v[key] ; found {
		return true
	}
	return false
}

func NewConnMap() *ConnMap {
	return &ConnMap{
		v: make(map[string]Conn),
	}
}