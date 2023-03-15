package isis

import (
	"container/heap"
	"context"
	"isis-total-order-multicast/config"
	"isis-total-order-multicast/gen-go/multicast/rpc"
	"isis-total-order-multicast/storage"
	"log"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Isis struct {
	peers map[string]*rpc.MulticastServiceClient
	id    string

	// global isis mgr lock
	lock sync.Mutex

	// chan for the peer deletion task
	delChan chan string

	// msg store
	storage storage.Storage

	// for the algo
	counter     int64
	counterLock sync.Mutex
	proposedSeq int64
	proposeLock sync.Mutex

	// hold back queue structure
	holdBackQueue HoldBackQueue
	holdBackMap   map[string]*QueueElem
	queueLock     sync.RWMutex
}

func NewIsis(config config.NodeConfig) *Isis {
	s := storage.NewFileStorage(config.StorageFilename)
	return &Isis{
		peers:         map[string]*rpc.MulticastServiceClient{},
		id:            config.Id,
		counter:       0,
		counterLock:   sync.Mutex{},
		lock:          sync.Mutex{},
		delChan:       make(chan string, 10),
		storage:       s,
		proposedSeq:   0,
		proposeLock:   sync.Mutex{},
		holdBackQueue: HoldBackQueue{},
		holdBackMap:   map[string]*QueueElem{},
		queueLock:     sync.RWMutex{},
	}
}

func (is *Isis) AddPeer(nodeId string, nodeCli *rpc.MulticastServiceClient) {
	is.lock.Lock()
	is.peers[nodeId] = nodeCli
	is.lock.Unlock()
}

func (is *Isis) RunDelTask() {
	for {
		nodeId := <-is.delChan
		is.lock.Lock()
		delete(is.peers, nodeId)
		is.lock.Unlock()
	}
}

func (is *Isis) TOMulticast(ctx context.Context, msg string) time.Time {
	is.lock.Lock()
	defer is.lock.Unlock()

	var wg sync.WaitGroup
	var lk sync.Mutex

	// multicast
	count := is.count()
	agreedSeq := int64(0)
	decisionMaker := is.id
	for node, cli := range is.peers {
		wg.Add(1)
		go func(node string, cli *rpc.MulticastServiceClient) {
			defer wg.Done()
			resp, err := cli.SendData(ctx, &rpc.SendDataRequest{
				Msg:    msg,
				MsgId:  count,
				Sender: is.id,
			})
			if err != nil {
				is.delChan <- node
				log.Printf("%v crashed: %v\n", node, err)
				return
			}
			lk.Lock()
			if resp.ProposedSeq > agreedSeq || (resp.ProposedSeq == agreedSeq && node > decisionMaker) {
				agreedSeq = resp.ProposedSeq
				decisionMaker = node
			}
			lk.Unlock()
		}(node, cli)
	}
	wg.Wait()

	// re-multicast after collecting all proposed seq
	endTime := time.Now()
	timeLock := sync.Mutex{}
	for node, cli := range is.peers {
		wg.Add(1)
		go func(node string, cli *rpc.MulticastServiceClient) {
			defer wg.Done()
			resp, err := cli.SendSeq(ctx, &rpc.SendSeqRequest{
				MsgId:         count,
				Sender:        is.id,
				AgreedSeq:     agreedSeq,
				DecisionMaker: decisionMaker,
			})
			if err != nil {
				is.delChan <- node
				log.Printf("%v crashed: %v\n", node, err)
				return
			}
			deliverTime, _ := time.Parse(resp.DeliverTime, "2006-01-02 15:04:05")
			timeLock.Lock()
			if endTime.Before(deliverTime) {
				endTime = deliverTime
			}
			timeLock.Unlock()
		}(node, cli)
	}
	wg.Wait()
	return endTime
}

func (is *Isis) Receive(msg string, msgId int64, sender string) int64 {
	propose := is.propose()
	elem := &QueueElem{
		Msg:       msg,
		MsgId:     msgId,
		Sender:    sender,
		Proposer:  is.id,
		Seq:       propose,
		Delivered: false,
	}
	is.queueLock.Lock()
	elemKey := elem.Sender + ":" + strconv.Itoa(int(elem.MsgId))
	is.holdBackMap[elemKey] = elem
	heap.Push(&is.holdBackQueue, elem)
	is.queueLock.Unlock()
	return propose
}

func (is *Isis) Deliver(msgId, seq int64, sender, proposer string) {
	is.updatePropose(seq)
	is.queueLock.Lock()
	elemKey := sender + ":" + strconv.Itoa(int(msgId))
	elem, ok := is.holdBackMap[elemKey]
	if !ok {
		panic(is.holdBackMap)
	}
	elem.Delivered = true
	elem.Seq = seq
	elem.Proposer = proposer
	sort.Stable(is.holdBackQueue)
	for cur, ok := is.holdBackQueue.Top(); ok && cur.(*QueueElem).Delivered; cur, ok = is.holdBackQueue.Top() {
		elem := heap.Pop(&is.holdBackQueue).(*QueueElem)
		_ = is.storage.Apply(elem.Msg)
	}
	is.queueLock.Unlock()
}

func (is *Isis) IsDelivered(msgId int64, nodeId string) bool {
	is.queueLock.RLock()
	defer is.queueLock.RUnlock()

	elemKey := nodeId + ":" + strconv.Itoa(int(msgId))
	elem, ok := is.holdBackMap[elemKey]
	if ok && elem.Delivered == true {
		return true
	}
	return false
}

func (is *Isis) BMulticast(ctx context.Context, msgId, seq int64, sender, proposer string) {
	is.lock.Lock()
	defer is.lock.Unlock()

	var wg sync.WaitGroup
	for node, cli := range is.peers {
		wg.Add(1)
		go func(node string, cli *rpc.MulticastServiceClient) {
			defer wg.Done()
			_, err := cli.SendSeq(ctx, &rpc.SendSeqRequest{
				MsgId:         msgId,
				Sender:        sender,
				AgreedSeq:     seq,
				DecisionMaker: proposer,
			})
			if err != nil {
				is.delChan <- node
				log.Printf("%v crashed: %v\n", node, err)
				return
			}
		}(node, cli)
	}
	wg.Wait()
}

func (is *Isis) count() int64 {
	return atomic.AddInt64(&is.counter, 1)
}

func (is *Isis) propose() int64 {
	return atomic.AddInt64(&is.proposedSeq, 1)
}

func (is *Isis) updatePropose(seq int64) {
	is.proposeLock.Lock()
	defer is.proposeLock.Unlock()
	if seq > is.proposedSeq {
		is.proposedSeq = seq
	}
}
