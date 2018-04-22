package priority_queue

import (
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	// create queue
	q := New(10)

	// push item
	i := Item{
		Value:    10,
		Priority: HighPriority,
	}
	q.Push(i)

	// get item
	i = q.Pop()
	if i.Value != 10 {
		t.Fatal("pq not working")
	}
}
