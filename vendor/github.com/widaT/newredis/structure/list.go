package structure

import (
	"fmt"
	"strings"
)


type Literator interface {
	// Next returns true if the iterator contains subsequent elements
	// and advances its state to the next element if that is possible.
	Next() (ok bool)
	// Value returns the current value.
	Value() []byte
	Key()  int
	// Close this iterator to reap resources associated with it.  While not
	// strictly required, it will provide extra hints for the garbage collector.
	Close()
}


type Liter struct {
	current *element
	list    * List
	index   int
	value   []byte
}


func (i Liter) Key() int {
	return i.index
}

func (i Liter) Value() []byte{
	return i.value
}

func (i *Liter) Next() bool {
	if !i.current.hasNext() {
		return false
	}
	i.current = i.current.next
	i.value = i.current.value
	i.index ++
	return true
}

func (i *Liter) Close() {
	i.index=0
	i.value = nil
	i.current = nil
	i.list = nil
}

type List struct {
	first *element
	last  *element
	size  int
}

type element struct {
	value []byte
	prev  *element
	next  *element
}

func (e *element ) hasNext() bool  {
	return e.next != nil
}


// New instantiates a new empty list
func NewList() *List {
	return &List{}
}

// Iterator returns an Iterator that will go through all elements s.
func (list *List) Literator() Literator {
	return &Liter{
		index:0,
		current: list.first,
		list:    list,
	}
}



// Seek returns a bidirectional iterator starting with the first element whose
// key is greater or equal to key; otherwise, a nil iterator is returned.
func (list *List) Seek(index int) Literator {
	current := list.GetIndex(index)
	if current == nil {
		return nil
	}

	return &Liter{
		index:index,
		current: current,
		list:    list,
		value:   current.value,
	}
}

func (list*List)GetIndex(index int) *element  {
	if !list.withinRange(index) {
		return nil
	}
	// determine traveral direction, last to first or first to last
	if list.size-index < index {
		element := list.last
		for e := list.size - 1; e != index; e, element = e-1, element.prev {
		}
		return element
	}
	element := list.first
	for e := 0; e != index; e, element = e+1, element.next {
	}
	return element
}

// Add appends a value (one or more) at the end of the list (same as Append())
func (list *List) Add(values ...[]byte) {
	for _, value := range values {
		newElement := &element{value: value, prev: list.last}
		if list.size == 0 {
			list.first = newElement
			list.last = newElement
		} else {
			list.last.next = newElement
			list.last = newElement
		}
		list.size++
	}
}

// Append appends a value (one or more) at the end of the list (same as Add())
func (list *List) Append(values ...[]byte) {
	list.Add(values...)
}

func (list *List) Lpush(values ...[]byte)  int {
	for v := 0; v <= len(values) - 1; v++ {
		newElement := &element{value: values[v], next: list.first}
		if list.size == 0 {
			list.first = newElement
			list.last = newElement
		} else {
			list.first.prev = newElement
			list.first = newElement
		}
		list.size++
	}
	return list.size
}

func (list *List) Rpush(values ...[]byte)  int {
	list.Add(values...)
	return list.size
}

func (list *List) Lpop() (value []byte) {
	value, _ = list.Get(0)
	list.Remove(0)
	return
}


func (list *List) Rpop() (value []byte) {
	value, _ = list.Get(list.size -1)
	list.Remove(list.size-1)
	return
}

// Prepend prepends a values (or more)
func (list *List) Prepend(values ...[]byte) {
	// in reverse to keep passed order i.e. ["c","d"] -> Prepend(["a","b"]) -> ["a","b","c",d"]
	for v := len(values) - 1; v >= 0; v-- {
		newElement := &element{value: values[v], next: list.first}
		if list.size == 0 {
			list.first = newElement
			list.last = newElement
		} else {
			list.first.prev = newElement
			list.first = newElement
		}
		list.size++
	}
}

// Get returns the element at index.
// Second return parameter is true if index is within bounds of the array and array is not empty, otherwise false.
func (list *List) Get(index int) ([]byte, bool) {

	if !list.withinRange(index) {
		return nil, false
	}

	// determine traveral direction, last to first or first to last
	if list.size-index < index {
		element := list.last
		for e := list.size - 1; e != index; e, element = e-1, element.prev {
		}
		return element.value, true
	}
	element := list.first
	for e := 0; e != index; e, element = e+1, element.next {
	}
	return element.value, true
}



// Remove removes one or more elements from the list with the supplied indices.
func (list *List) Remove(index int) {

	if !list.withinRange(index) {
		return
	}

	if list.size == 1 {
		list.Clear()
		return
	}

	var element *element
	// determine traversal direction, last to first or first to last
	if list.size-index < index {
		element = list.last
		for e := list.size - 1; e != index; e, element = e-1, element.prev {
		}
	} else {
		element = list.first
		for e := 0; e != index; e, element = e+1, element.next {
		}
	}

	if element == list.first {
		list.first = element.next
	}
	if element == list.last {
		list.last = element.prev
	}
	if element.prev != nil {
		element.prev.next = element.next
	}
	if element.next != nil {
		element.next.prev = element.prev
	}

	element = nil

	list.size--
}


// Values returns all elements in the list.
func (list *List) Values() [][]byte {
	values := make([][]byte, list.size, list.size)
	for e, element := 0, list.first; element != nil; e, element = e+1, element.next {
		values[e] = element.value
	}
	return values
}

// Empty returns true if list does not contain any elements.
func (list *List) Empty() bool {
	return list.size == 0
}

// Size returns number of elements within the list.
func (list *List) Size() int {
	return list.size
}

// Clear removes all elements from the list.
func (list *List) Clear() {
	list.size = 0
	list.first = nil
	list.last = nil
}



// Swap swaps values of two elements at the given indices.
func (list *List) Swap(i, j int) {
	if list.withinRange(i) && list.withinRange(j) && i != j {
		var element1, element2 *element
		for e, currentElement := 0, list.first; element1 == nil || element2 == nil; e, currentElement = e+1, currentElement.next {
			switch e {
			case i:
				element1 = currentElement
			case j:
				element2 = currentElement
			}
		}
		element1.value, element2.value = element2.value, element1.value
	}
}

// Insert inserts values at specified index position shifting the value at that position (if any) and any subsequent elements to the right.
// Does not do anything if position is negative or bigger than list's size
// Note: position equal to list's size is valid, i.e. append.
func (list *List) Insert(index int, values ...[]byte) {

	if !list.withinRange(index) {
		// Append
		if index == list.size {
			list.Add(values...)
		}
		return
	}

	list.size += len(values)

	var beforeElement *element
	var foundElement *element
	// determine traversal direction, last to first or first to last
	if list.size-index < index {
		foundElement = list.last
		for e := list.size - 1; e != index; e, foundElement = e-1, foundElement.prev {
			beforeElement = foundElement.prev
		}
	} else {
		foundElement = list.first
		for e := 0; e != index; e, foundElement = e+1, foundElement.next {
			beforeElement = foundElement
		}
	}

	if foundElement == list.first {
		oldNextElement := list.first
		for i, value := range values {
			newElement := &element{value: value}
			if i == 0 {
				list.first = newElement
			} else {
				newElement.prev = beforeElement
				beforeElement.next = newElement
			}
			beforeElement = newElement
		}
		oldNextElement.prev = beforeElement
		beforeElement.next = oldNextElement
	} else {
		oldNextElement := beforeElement.next
		for _, value := range values {
			newElement := &element{value: value}
			newElement.prev = beforeElement
			beforeElement.next = newElement
			beforeElement = newElement
		}
		oldNextElement.prev = beforeElement
		beforeElement.next = oldNextElement
	}
}

// String returns a string representation of container
func (list *List) String() string {
	str := "DoublyLinkedList\n"
	values := []string{}
	for element := list.first; element != nil; element = element.next {
		values = append(values, fmt.Sprintf("%v", string(element.value)))
	}
	str += strings.Join(values, ", ")
	return str
}

// Check that the index is within bounds of the list
func (list *List) withinRange(index int) bool {
	return index >= 0 && index < list.size
}
