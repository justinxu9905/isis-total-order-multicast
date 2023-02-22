package multicast

import (
	"isis-total-order-multicast/gen-go/multicast/rpc"
	"isis-total-order-multicast/storage"
	"os"
	"sync"
)

type Mgr struct {
	peers   map[string]*rpc.MulticastServiceClient
	id        string

	// global multicast mgr lock
	lock sync.Mutex

	// chan for the peer deletion task
	delChan      chan string

	// msg store
	storage       *storage.Storage

	// for the algo
	counter       int64
	counterLock   sync.Mutex
	proposedSeq   int64
	proposeLock   sync.Mutex

	// hold back queue structure
	holdBackQueue       HoldBackQueue
	holdBackMap         map[string]*QueueElem
	queueLock     sync.RWMutex

	// logger
	logger        sync.Mutex
	logFile       *os.File
}