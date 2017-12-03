package priority_queue

//go:generate stringer -type=Priority

// Priority of signing request
type Priority int

const (
	// UnknownPriority represents an Unknown Priority signing request
	UnknownPriority Priority = iota
	// LowPriority represents an Low Priority signing request
	LowPriority
	// MediumPriority represents an Medium Priority signing request
	MediumPriority
	// HighPriority represents an High Priority signing request
	HighPriority
)

type Item struct {
	Value    interface{}
	Priority Priority
}

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

// Push reads a file and stores it at temporary location so that it can be
// processed later without consuming memory. The function returns a tracking
// id or error.
func (q *PriorityQueue) Push(i Item) error {
	switch i.Priority {
	case HighPriority:
		q.high <- i
	case MediumPriority:
		q.medium <- i
	case LowPriority:
		q.low <- i
	}

	return nil
}

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
