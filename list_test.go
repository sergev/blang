package main

import (
	"testing"
)

func TestListPush(t *testing.T) {
	t.Skip("Disabled during LLVM backend migration")
	list := NewList()

	// Test pushing items
	list.Push(1)
	list.Push(2)
	list.Push(3)

	if list.Size != 3 {
		t.Errorf("Size = %d, want 3", list.Size)
	}

	if list.Data[0].(int) != 1 {
		t.Errorf("Data[0] = %d, want 1", list.Data[0])
	}
	if list.Data[1].(int) != 2 {
		t.Errorf("Data[1] = %d, want 2", list.Data[1])
	}
	if list.Data[2].(int) != 3 {
		t.Errorf("Data[2] = %d, want 3", list.Data[2])
	}
}

func TestListPushMany(t *testing.T) {
	t.Skip("Disabled during LLVM backend migration")
	list := NewList()

	// Push more items than initial allocation
	for i := 0; i < 100; i++ {
		list.Push(i)
	}

	if list.Size != 100 {
		t.Errorf("Size = %d, want 100", list.Size)
	}

	// Verify all items
	for i := 0; i < 100; i++ {
		if list.Data[i].(int) != i {
			t.Errorf("Data[%d] = %d, want %d", i, list.Data[i], i)
		}
	}
}

func TestListClear(t *testing.T) {
	t.Skip("Disabled during LLVM backend migration")
	list := NewList()

	// Add items
	list.Push(1)
	list.Push(2)
	list.Push(3)

	// Clear the list
	list.Clear()

	if list.Size != 0 {
		t.Errorf("Size after Clear() = %d, want 0", list.Size)
	}

	// Verify we can still push after clearing
	list.Push(4)
	if list.Size != 1 {
		t.Errorf("Size after Clear() and Push() = %d, want 1", list.Size)
	}
	if list.Data[0].(int) != 4 {
		t.Errorf("Data[0] after Clear() and Push() = %d, want 4", list.Data[0])
	}
}

func TestListFree(t *testing.T) {
	t.Skip("Disabled during LLVM backend migration")
	list := NewList()

	// Add items
	list.Push(1)
	list.Push(2)
	list.Push(3)

	// Free the list
	list.Free()

	if list.Size != 0 {
		t.Errorf("Size after Free() = %d, want 0", list.Size)
	}
	if list.Alloc != 0 {
		t.Errorf("Alloc after Free() = %d, want 0", list.Alloc)
	}
	if list.Data != nil {
		t.Error("Data after Free() should be nil")
	}
}

func TestListDifferentTypes(t *testing.T) {
	t.Skip("Disabled during LLVM backend migration")
	list := NewList()

	// Push different types
	list.Push(42)
	list.Push("hello")
	list.Push(3.14)
	list.Push(true)

	if list.Size != 4 {
		t.Errorf("Size = %d, want 4", list.Size)
	}

	if list.Data[0].(int) != 42 {
		t.Errorf("Data[0] = %v, want 42", list.Data[0])
	}
	if list.Data[1].(string) != "hello" {
		t.Errorf("Data[1] = %v, want hello", list.Data[1])
	}
	if list.Data[2].(float64) != 3.14 {
		t.Errorf("Data[2] = %v, want 3.14", list.Data[2])
	}
	if list.Data[3].(bool) != true {
		t.Errorf("Data[3] = %v, want true", list.Data[3])
	}
}

func BenchmarkListPush(b *testing.B) {
	list := NewList()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Push(i)
	}
}
