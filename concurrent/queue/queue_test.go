package queue

import (
	"runtime"
	"sort"
	"sync"
	"testing"
)

const (
	kGoRoutineNum = 10
	kPushingNum   = 500000
	kBufSz        = kGoRoutineNum * kPushingNum
)

var out *testing.T
var wg sync.WaitGroup
var lfq = NewQueue[int]()
var popBuf [kGoRoutineNum][]int

func TestQueue(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	out = t
	// init popBuf
	for i := 0; i != kGoRoutineNum; i++ {
		popBuf[i] = make([]int, 0, kBufSz)
	}

	// Push() simultaneously
	wg.Add(kGoRoutineNum)
	for i := 0; i != kGoRoutineNum; i++ {
		go push()
	}
	wg.Wait()
	// Pop() simultaneously
	wg.Add(kGoRoutineNum)
	for i := 0; i != kGoRoutineNum; i++ {
		go popOnly()
	}
	wg.Wait()

	// Push() and Pop() simultaneously
	wg.Add(kGoRoutineNum * 2)
	for i := 0; i != kGoRoutineNum; i++ {
		go push()
		go popWhilePushing(i)
	}
	wg.Wait()
	// Verification
	resultBuf := popBuf[0]
	for i := 1; i != kGoRoutineNum; i++ {
		resultBuf = append(resultBuf, popBuf[i]...)
	}
	// in case there are some elements left in the queue
	for v, ok := lfq.Pop(); ok; v, ok = lfq.Pop() {
		resultBuf = append(resultBuf, v)
	}
	sort.Ints(resultBuf)
	for i := 0; i != kPushingNum; i++ {
		for j := 0; j != kGoRoutineNum; j++ {
			if resultBuf[(i*kGoRoutineNum)+j] != i {
				t.Error("Invalid result:", i, j, resultBuf[(i*kGoRoutineNum)+j])
			}
		}
	}
}

func push() {
	for i := 0; i != kPushingNum; i++ {
		lfq.Push(i)
	}
	wg.Done()
}

func popOnly() {
	for i := 0; i != kPushingNum; i++ {
		_, ok := lfq.Pop()
		if !ok {
			out.Error("Should never be nil!")
		}
	}
	wg.Done()
}

func popWhilePushing(n int) {
	for i := 0; i != kPushingNum*2; i++ {
		v, ok := lfq.Pop()
		if ok {
			popBuf[n] = append(popBuf[n], v)
		}
	}
	wg.Done()
}
