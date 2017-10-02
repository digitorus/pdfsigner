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
		value: job{
			ID:       generateID(),
			fileName: "myfile",
		},
		priority: HighPriority,
	}

	q.Push(i)
	log.Println(q.Pop())
}

func generateID() string {
	guid := xid.New()
	return guid.String()
}

//// Queues exposes an transparent interface to the sign queue, all clients should
//// implement this interface.
////
//// The crypto.Signer map can contain multiple singers as defined in the config,
//// a Singer implementation can be a private key or PKCS#11 device.
//var Queues map[string]priorityQueue

//
//// job is the internal job specification
//// contains information required to schedule a document to be processed
//// by one of the signers.
//type job struct {
//	ID   string
//	file string // tmp or real file location
//}
//
//// create a unique id that can be used by a client to obtain the document or
//// current state of the Item
//guid := xid.New()
//jobID := guid.String()
//
//job := job{
//ID:   jobID,
//file: fileName,
//}
