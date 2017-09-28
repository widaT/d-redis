package structure

import (
	"math/rand"
	"fmt"
)

const p = 0.25

const DefaultMaxLevel = 32


type node struct {
	forward    []*node
	backward   *node
	key  float64
	value  string
}

func (n *node) next() *node {
	if len(n.forward) == 0 {
		return nil
	}
	return n.forward[0]
}

// previous returns the previous node in the skip list containing n.
func (n *node) previous() *node {
	return n.backward
}

// hasNext returns true if n has a next node.
func (n *node) hasNext() bool {
	return n.next() != nil
}

// hasPrevious returns true if n has a previous node.
func (n *node) hasPrevious() bool {
	return n.previous() != nil
}

// A SkipList is a map-like data structure that maintains an ordered
// collection of key-value pairs. Insertion, lookup, and deletion are
// all O(log n) operations. A SkipList can efficiently store up to
// 2^MaxLevel items.
//
// To iterate over a skip list (where s is a
// *SkipList):
//
//	for i := s.Iterator(); i.Next(); {
//		// do something with i.Key() and i.Value()
//	}
type SkipList struct {
	header   *node
	footer   *node
	length   int
	MaxLevel int
}

// Len returns the length of s.
func (s *SkipList) Len() int {
	return s.length
}

// Iterator is an interface that you can use to iterate through the
// skip list (in its entirety or fragments). For an use example, see
// the documentation of SkipList.
//
// Key and Value return the key and the value of the current node.
type Iterator interface {
	// Next returns true if the iterator contains subsequent elements
	// and advances its state to the next element if that is possible.
	Next() (ok bool)

	// Previous returns true if the iterator contains previous elements
	// and rewinds its state to the previous element if that is possible.
	Previous() (ok bool)
	// Key returns the current key.
	Key() float64
	// Value returns the current value.
	Value() string
	// Seek reduces iterative seek costs for searching forward into the Skip List
	// by remarking the range of keys over which it has scanned before.  If the
	// requested key occurs prior to the point, the Skip List will start searching
	// as a safeguard.  It returns true if the key is within the known range of
	// the list.
	Seek(key float64) (ok bool)
	// Close this iterator to reap resources associated with it.  While not
	// strictly required, it will provide extra hints for the garbage collector.
	Close()
}

type iter struct {
	current *node
	key     float64
	list    *SkipList
	value   string
}

func (i iter) Key() float64 {
	return i.key
}

func (i iter) Value() string {
	return i.value
}

func (i *iter) Next() bool {
	if !i.current.hasNext() {
		return false
	}

	i.current = i.current.next()
	i.key = i.current.key
	i.value = i.current.value

	return true
}



func (i *iter) Previous() bool {
	if !i.current.hasPrevious() {
		return false
	}

	i.current = i.current.previous()
	i.key = i.current.key
	i.value = i.current.value

	return true
}

func (i *iter) Seek(key float64) (ok bool) {
	current := i.current
	list := i.list

	// If the existing iterator outside of the known key range, we should set the
	// position back to the beginning of the list.
	if current == nil {
		current = list.header
	}

	// If the target key occurs before the current key, we cannot take advantage
	// of the heretofore spent traversal cost to find it; resetting back to the
	// beginning is the safest choice.
	if  key < current.key {
		current = list.header
	}

	// We should back up to the so that we can seek to our present value if that
	// is requested for whatever reason.
	if current.backward == nil {
		current = list.header
	} else {
		current = current.backward
	}

	current = list.getPath(current, nil, key,"")

	if current == nil {
		return
	}
	i.current = current
	i.key = current.key
	i.value = current.value
	for pre := current.previous();pre.key == key;pre = pre.previous(){
		i.current = pre
		i.key = pre.key
		i.value = pre.value
	}
	return true
}

func (i *iter) Close() {
	i.key = 0
	i.value = ""
	i.current = nil
	i.list = nil
}

type rangeIterator struct {
	iter
	upperLimit float64
	lowerLimit float64
}

func (i *rangeIterator) Next() bool {
	if !i.current.hasNext() {
		return false
	}
	next := i.current.next()
	if next.key > i.upperLimit {
		return false
	}

	i.current = next
	i.key = next.key
	i.value = next.value
	return true
}


func (i *rangeIterator) Previous() bool {
	if !i.current.hasPrevious() {
		return false
	}

	previous := i.current.previous()

	if  previous.key < i.lowerLimit  {
		return false
	}

	i.current = i.current.previous()
	i.key = i.current.key
	i.value = i.current.value
	return true
}

func (i *rangeIterator) Seek(key float64) (ok bool) {
	if key <i.lowerLimit {
		return
	} else if  key < i.upperLimit {
		return
	}

	return i.iter.Seek(key)
}

func (i *rangeIterator) Close() {
	i.iter.Close()
	i.upperLimit = 0
	i.lowerLimit = 0
}

type indexRangeIterator struct {
	iter
	index int
	upperLimit int
	lowerLimit int
}
func (i *indexRangeIterator) Close() {
	i.iter.Close()
	i.index =0
	i.upperLimit = 0
	i.lowerLimit = 0
}
func (i *indexRangeIterator) Next() bool {
	if !i.current.hasNext() {
		return false
	}
	if i.index > i.upperLimit {
		return false
	}
	i.current = i.current.next()
	i.key = i.current.key
	i.value = i.current.value
	i.index ++
	return true
}

func (i *indexRangeIterator) Previous() bool {
	if !i.current.hasPrevious() {
		return false
	}
	i.index ++
	if i.index <= i.lowerLimit {
		return false
	}

	i.current = i.current.previous()
	i.key = i.current.key
	i.value = i.current.value
	return true
}

// Iterator returns an Iterator that will go through all elements s.
func (s *SkipList) Iterator() Iterator {
	return &iter{
		current: s.header,
		list:    s,
	}
}


// Seek returns a bidirectional iterator starting with the first element whose
// key is greater or equal to key; otherwise, a nil iterator is returned.
func (s *SkipList) Seek(key float64) Iterator {
	current := s.getPath(s.header, nil, key,"")

	fmt.Println(current)
	if current == nil {
		return nil
	}

	i := iter{
		current: current,
		key:     current.key,
		list:    s,
		value:   current.value,
	}

	return &i
}

// SeekToFirst returns a bidirectional iterator starting from the first element
// in the list if the list is populated; otherwise, a nil iterator is returned.
func (s *SkipList) SeekToFirst() Iterator {
	if s.length == 0 {
		return nil
	}

	current := s.header.next()

	return &iter{
		current: current,
		key:     current.key,
		list:    s,
		value:   current.value,
	}
}

// SeekToLast returns a bidirectional iterator starting from the last element
// in the list if the list is populated; otherwise, a nil iterator is returned.
func (s *SkipList) SeekToLast() Iterator {
	current := s.footer
	if current == nil {
		return nil
	}

	return &iter{
		current: current,
		key:     current.key,
		list:    s,
		value:   current.value,
	}
}

// Range returns an iterator that will go through all the
// elements of the skip list that are greater or equal than from, but
// less than to.
func (s *SkipList) Range(from, to float64) Iterator {
	start := s.getPath(s.header, nil, from,"")
	return &rangeIterator{
		iter: iter{
			current: &node{
				forward:  []*node{start},
				backward: start,
			},
			list: s,
		},
		upperLimit: to,
		lowerLimit: from,
	}
}


func (s *SkipList) IndexRange(l, u int) Iterator {
	n := s.header
	for  i := 0;i <= l;i++ {
		n = n.next()
	}
	return &indexRangeIterator{
		iter: iter{
			current: &node{
				forward:  []*node{n},
				backward: n,
			},
			list: s,
		},
		index:l,
		upperLimit: u,
		lowerLimit: l,
	}
}

func (s *SkipList) level() int {
	return len(s.header.forward) - 1
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func (s *SkipList) effectiveMaxLevel() int {
	return maxInt(s.level(), s.MaxLevel)
}

// Returns a new random level.
func (s SkipList) randomLevel() (n int) {
	for n = 0; n < s.effectiveMaxLevel() && rand.Float64() < p; n++ {
	}
	return
}

// Get returns the value associated with key from s (nil if the key is
// not present in s). The second return value is true when the key is
// present.
func (s *SkipList) Get(key float64) (value string, ok bool) {
	candidate := s.getPath(s.header, nil, key,"")

	if candidate == nil || candidate.key != key {
		return "", false
	}

	return candidate.value, true
}

// GetGreaterOrEqual finds the node whose key is greater than or equal
// to min. It returns its value, its actual key, and whether such a
// node is present in the skip list.
func (s *SkipList) GetGreaterOrEqual(min float64) (actualKey float64, value string, ok bool) {
	candidate := s.getPath(s.header, nil, min,"")

	if candidate != nil {
		return candidate.key, candidate.value, true
	}
	return 0, "", false
}

// getPath populates update with nodes that constitute the path to the
// node that may contain key. The candidate node will be returned. If
// update is nil, it will be left alone (the candidate node will still
// be returned). If update is not nil, but it doesn't have enough
// slots for all the nodes in the path, getPath will panic.
func (s *SkipList) getPath(current *node, update []*node, key float64,value string) *node {
	depth := len(current.forward) - 1

	for i := depth; i >= 0; i-- {
		for current.forward[i] != nil && ( current.forward[i].key <  key || ( current.forward[i].key ==  key && current.forward[i].value < value))  {
			current = current.forward[i]
		}
		if update != nil {
			update[i] = current
		}
	}
	return current.next()
}

// Sets set the value associated with key in s.
func (s *SkipList) Set(key float64, value string) {
	// s.level starts from 0, so we need to allocate one.
	update := make([]*node, s.level()+1, s.effectiveMaxLevel()+1)
	candidate := s.getPath(s.header, update, key,value)

	if candidate != nil && candidate.key == key && candidate.value == value {
		return
	}

	newLevel := s.randomLevel()

	if currentLevel := s.level(); newLevel > currentLevel {
		// there are no pointers for the higher levels in
		// update. Header should be there. Also add higher
		// level links to the header.
		for i := currentLevel + 1; i <= newLevel; i++ {
			update = append(update, s.header)
			s.header.forward = append(s.header.forward, nil)
		}
	}

	newNode := &node{
		forward: make([]*node, newLevel+1, s.effectiveMaxLevel()+1),
		key:     key,
		value:   value,
	}

	previous := update[0]
	newNode.backward = previous


	for i := 0; i <= newLevel; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}

	s.length++

	if newNode.forward[0] != nil {
		if newNode.forward[0].backward != newNode {
			newNode.forward[0].backward = newNode
		}
	}

	if s.footer == nil || s.footer.key < key {
		s.footer = newNode
	}
}

// Delete removes the node with the given key.
//
// It returns the old value and whether the node was present.
func (s *SkipList) DeleteAll(key float64)  bool {
	update := make([]*node, s.level()+1, s.effectiveMaxLevel())
	candidate := s.getPath(s.header, update, key,"")


	if candidate == nil || candidate.key != key {
		return  false
	}

	for next := candidate.next();next.key == key;next = candidate.next() {
		s.Delete(next.key,next.value)
	}

	for pre := candidate.previous();pre.key == key;pre = candidate.previous() {
		s.Delete(pre.key,pre.value)
	}

	if candidate.key == key {
		s.Delete(candidate.key,candidate.value)
	}

	return true
}


// Delete removes the node with the given key.
//
// It returns the old value and whether the node was present.
func (s *SkipList) Delete(key float64,value string) (ok bool) {
	update := make([]*node, s.level()+1, s.effectiveMaxLevel())
	candidate := s.getPath(s.header, update, key,value)

	if candidate == nil || candidate.value != value {
		return false
	}

	previous := candidate.backward
	if s.footer == candidate {
		s.footer = previous
	}

	next := candidate.next()
	if next != nil {
		next.backward = previous
	}

	for i := 0; i <= s.level() && update[i].forward[i] == candidate; i++ {
		update[i].forward[i] = candidate.forward[i]
	}

	for s.level() > 0 && s.header.forward[s.level()] == nil {
		s.header.forward = s.header.forward[:s.level()]
	}
	s.length--

	return true
}


// Ordered is an interface which can be linearly ordered by the
// LessThan method, whereby this instance is deemed to be less than
// other. Additionally, Ordered instances should behave properly when
// compared using == and !=.
type Ordered interface {
	LessThan(other Ordered) bool
}

// New returns a new SkipList.
//
// Its keys must implement the Ordered interface.
func NewSkipList() *SkipList {
	return &SkipList{
		header: &node{
			forward: []*node{nil},
		},
		MaxLevel: DefaultMaxLevel,
	}
}