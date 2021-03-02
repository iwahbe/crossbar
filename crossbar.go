package crossbar

import (
	"sync"
)

type Crossbar struct {
	channels []chan int
	P        int
	barrier  Barrier
}

func (self *Crossbar) Setup(P int) {
	size := P * P
	self.P = P
	self.channels = make([]chan int, size)
	self.barrier = NewBarrier(P)
	for i := 0; i < size; i++ {
		self.channels[i] = make(chan int)
	}
}

func (self *Crossbar) Receive(receiver, sender int) int {
	return <-self.channels[(receiver*self.P)+sender]
}

func (self *Crossbar) Send(sender, receiver, value int) {
	self.channels[(receiver*self.P)+sender] <- value
}

type Node struct {
	crossbar *Crossbar
	ProcNum  int
}

func (self *Crossbar) Node(id int) Node {
	var node Node
	node.crossbar = self
	node.ProcNum = id
	return node
}

func (self *Node) Send(receiver, value int) {
	self.crossbar.Send(self.ProcNum, receiver, value)
}

func (self *Node) Receive(sender int) int {
	return self.crossbar.Receive(self.ProcNum, sender)
}

func (self *Node) Synchronize() {
	self.crossbar.barrier.Rendezvous()
}

// Cannot be coppied
type Barrier struct {
	cond      *sync.Cond
	waitCount int
	maxCount  int
}

func NewBarrier(num int) Barrier {
	var mtx sync.Mutex
	var self Barrier
	self.cond = sync.NewCond(&mtx)
	self.waitCount = num
	self.maxCount = num
	return self
}

func (self *Barrier) Rendezvous() {
	self.cond.L.Lock()
	defer self.cond.L.Unlock()

	self.waitCount -= 1
	// The last one to rendezvous
	if self.waitCount == 0 {
		self.waitCount = self.maxCount
		self.cond.Broadcast()
	} else {
		self.cond.Wait() // Unlocks until done
	}
}
