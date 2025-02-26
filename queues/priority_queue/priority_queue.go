package priority_queue

import "errors"

//go:generate stringer -type=Priority

// Priority of signing request.
type Priority int

const (
	// UnknownPriority represents an Unknown Priority signing request.
	UnknownPriority Priority = iota
	// LowPriority represents an Low Priority signing request.
	LowPriority
	// MediumPriority represents an Medium Priority signing request.
	MediumPriority
	// HighPriority represents an High Priority signing request.
	HighPriority
)

// Item is used to push to and pop value with priority.
type Item struct {
	// Value represents any value
	Value interface{}
	// Priority represents priority of the signing reqeust
	Priority Priority
}

// PriorityQueue represents priority channels.
type PriorityQueue struct {
	high   chan Item
	medium chan Item
	low    chan Item
}

func New(size int) *PriorityQueue {
	q := PriorityQueue{
		high:   make(chan Item, size),
		medium: make(chan Item, size),
		low:    make(chan Item, size),
	}

	return &q
}

// Push adds an item to the priority queue.
func (q *PriorityQueue) Push(i Item) {
	switch i.Priority {
	case HighPriority:
		q.high <- i
	case MediumPriority:
		q.medium <- i
	case LowPriority:
		q.low <- i
	}
}

// Pop returns appropriate item from the priority queue.
func (q *PriorityQueue) Pop() Item {
	for {
		select {
		case Item := <-q.high:
			return Item
		default:
			select {
			case Item := <-q.medium:
				return Item
			case Item := <-q.low:
				return Item
			default:
				select {
				case Item := <-q.high:
					return Item
				case Item := <-q.medium:
					return Item
				case Item := <-q.low:
					return Item
				}
			}
		}
	}
}

// LenAll represents lengths of priority channels.
type LenAll struct {
	Low    int `json:"low"`
	Medium int `json:"medium"`
	High   int `json:"high"`
}

// Len returns length of the channel by priority.
func (q *PriorityQueue) Len(p Priority) (int, error) {
	switch p {
	case LowPriority:
		return len(q.low), nil
	case MediumPriority:
		return len(q.medium), nil
	case HighPriority:
		return len(q.high), nil
	}

	return -1, errors.New("wrong priority name")
}

// LenAll returns lengths of all priority channels.
func (q *PriorityQueue) LenAll() LenAll {
	return LenAll{len(q.low), len(q.medium), len(q.high)}
}
