package goctx

import (
	"context"
	"sync"
)

type Owner struct {
	ctx	 context.Context
	cancel	 context.CancelFunc
	canceled bool
	mux	 sync.Mutex
	wg	 sync.WaitGroup
}

type Worker struct {
	ctx	 context.Context
	cancel	 context.CancelFunc
	canceled *bool
	mux	 *sync.Mutex
	wg	 *sync.WaitGroup
}

func NewOwner() Owner {
	ctx, cancel := context.WithCancel(context.Background())
	return Owner{ctx:ctx, cancel:cancel, mux:sync.Mutex{}, wg:sync.WaitGroup{}}
}

func (self *Owner)NewWorker() Worker {
	ctx, _ := context.WithCancel(self.ctx)
	self.wg.Add(1)
	return Worker{ctx:ctx, cancel:self.cancel, mux:&self.mux,
					wg:&self.wg, canceled:&self.canceled}
}

func (self *Owner) Wait() {
	self.wg.Wait()
}

func (self *Owner) Lock() {
	self.mux.Lock()
}

func (self *Owner) Unlock() {
	self.mux.Unlock()
}

func (self *Owner) Done() {
	self.wg.Done()
}

func (self *Owner) Cancel() {
	self.mux.Lock()
	defer self.mux.Unlock()
	if self.canceled {
		return
	}

	self.canceled = true
	self.cancel()
}

func (self *Worker) Lock() {
	self.mux.Lock()
}

func (self *Worker) Unlock() {
	self.mux.Unlock()
}

func (self *Worker) Done() {
	self.wg.Done()
}

func (self *Worker) Cancel() {
	self.mux.Lock()
	defer self.mux.Unlock()
	if *self.canceled {
		return
	}

	*self.canceled = true
	self.cancel()
}

func (self *Worker) RecvCancel() <-chan struct{} {
	return self.ctx.Done()
}
