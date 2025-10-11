package main

// List is a dynamic array that can hold any type
type List struct {
	Alloc int           // allocated capacity
	Size  int           // current size
	Data  []interface{} // data storage
}

// NewList creates a new empty list
func NewList() *List {
	return &List{
		Alloc: 0,
		Size:  0,
		Data:  nil,
	}
}

// Push adds an item to the list
func (l *List) Push(item interface{}) {
	if l.Alloc == 0 {
		l.Alloc = 32
		l.Data = make([]interface{}, l.Alloc)
	} else if l.Alloc-l.Size < 2 {
		l.Alloc *= 2
		newData := make([]interface{}, l.Alloc)
		copy(newData, l.Data[:l.Size])
		l.Data = newData
	}
	l.Data[l.Size] = item
	l.Size++
}

// Clear resets the list size to 0
func (l *List) Clear() {
	for i := 0; i < l.Size; i++ {
		l.Data[i] = nil
	}
	l.Size = 0
}

// Free releases the list's memory
func (l *List) Free() {
	l.Data = nil
	l.Alloc = 0
	l.Size = 0
}
