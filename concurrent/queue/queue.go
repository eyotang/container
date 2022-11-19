/*
Package queue provides a fast, ring-buffer queue based on the version suggested by Dariusz Górecki.
Using this instead of other, simpler, queue implementations (slice+append or linked list) provides
substantial memory and time benefits, and fewer GC pauses.
The queue implemented here is as fast as it is for an additional reason: it is *not* thread-safe.
*/
package queue

import "sync"

// minQueueLen is smallest capacity that queue may have.
// Must be power of 2 for bitwise modulus: x % n == x & (n - 1).
const minQueueLen = 16

// Queue represents a single instance of the queue data structure.
type Queue[T comparable] struct {
	buf               []T
	head, tail, count int
	lock              sync.RWMutex
}

// NewQueue constructs and returns a new Queue.
func NewQueue[T comparable]() *Queue[T] {
	return &Queue[T]{
		buf: make([]T, minQueueLen),
	}
}

// Size returns the number of elements currently stored in the queue.
func (q *Queue[T]) Size() int {
	q.lock.RLock()
	count := q.count
	q.lock.RUnlock()
	return count
}

func (q *Queue[T]) Empty() bool {
	return q.Size() == 0
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
func (q *Queue[T]) resize() {
	newBuf := make([]T, q.count<<1)

	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

// Push puts an element on the end of the queue.
func (q *Queue[T]) Push(elem T) {
	q.lock.Lock()
	if q.count == len(q.buf) {
		q.resize()
	}

	q.buf[q.tail] = elem
	// bitwise modulus
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
	q.lock.Unlock()
}

// Peek returns the element at the head of the queue. This call panics
// if the queue is empty.
func (q *Queue[T]) Peek() T {
	q.lock.RLock()
	if q.count <= 0 {
		q.lock.RUnlock()
		panic("queue: Peek() called on empty queue")
	}
	v := q.buf[q.head]
	q.lock.RUnlock()
	return v
}

// Get returns the element at index i in the queue. If the index is
// invalid, the call will panic. This method accepts both positive and
// negative index values. Index 0 refers to the first element, and
// index -1 refers to the last.
func (q *Queue[T]) Get(i int) T {
	q.lock.RLock()
	// If indexing backwards, convert to positive index.
	if i < 0 {
		i += q.count
	}
	if i < 0 || i >= q.count {
		q.lock.RUnlock()
		panic("queue: Get() called with index out of range")
	}
	// bitwise modulus
	v := q.buf[(q.head+i)&(len(q.buf)-1)]
	q.lock.RUnlock()
	return v
}

// Pop removes and returns the element from the front of the queue. If the
// queue is empty, the call will panic.
func (q *Queue[T]) Pop() (T, bool) {
	q.lock.Lock()
	if q.count <= 0 {
		q.lock.Unlock()
		var v T
		return v, false
	}
	ret := q.buf[q.head]
	//q.buf[q.head] = nil
	// bitwise modulus
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	// Resize down if buffer 1/4 full.
	if len(q.buf) > minQueueLen && (q.count<<2) == len(q.buf) {
		q.resize()
	}
	q.lock.Unlock()
	return ret, true
}

func (q *Queue[T]) Index(val T) int {
	q.lock.RLock()
	if q.count <= 0 {
		return -1
	}
	idx := 0
	if q.tail > q.head {
		for i := q.head; i < q.tail; i++ {
			if q.buf[i] == val {
				q.lock.RUnlock()
				return idx
			}
			idx++
		}
	} else {
		for i := q.head; i < len(q.buf); i++ {
			if q.buf[i] == val {
				q.lock.RUnlock()
				return idx
			}
			idx++
		}
		for i := 0; i < q.tail; i++ {
			if q.buf[i] == val {
				q.lock.RUnlock()
				return idx
			}
			idx++
		}
	}
	q.lock.RUnlock()
	return -1
}
