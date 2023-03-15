package isis

type QueueElem struct {
	Msg       string
	MsgId     int64
	Sender    string
	Proposer  string
	Seq       int64
	Delivered bool
}

// max heap by default
type HoldBackQueue []*QueueElem

func (h HoldBackQueue) Len() int { return len(h) }

func (h HoldBackQueue) Less(i, j int) bool {
	if h[i].Seq != h[j].Seq {
		return h[i].Seq < h[j].Seq
	} else {
		return h[i].Proposer < h[j].Proposer
	}
}

func (h HoldBackQueue) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *HoldBackQueue) Pop() interface{} {
	old := *h
	n := h.Len()
	if n == 0 {
		return nil
	}
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *HoldBackQueue) Top() (interface{}, bool) {
	n := h.Len()
	if n == 0 {
		return nil, false
	}
	return (*h)[0], true
}

func (h *HoldBackQueue) Push(x interface{}) {
	*h = append(*h, x.(*QueueElem))
}
