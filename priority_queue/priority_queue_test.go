package priority_queue

import (
	"log"
	"testing"

	"github.com/rs/xid"
)

// job is the internal job specification
// contains information required to schedule a document to be processed
// by one of the signers.
type job struct {
	ID       string
	fileName string // tmp or real file location
}

// create a unique id that can be used by a client to obtain the document or
// current state of the Item
func TestPriorityQueue(t *testing.T) {
	q := NewPriorityQueue(10)
	i := Item{
		Value: job{
			ID:       generateID(),
			fileName: "myfile",
		},
		Priority: HighPriority,
	}

	q.Push(i)
	log.Println(q.Pop())
}

func generateID() string {
	guid := xid.New()
	return guid.String()
}

