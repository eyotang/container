package queue

// A Queue is a queue of element.
type Queue struct {
	// This is a queue, not a deque.
	// It is split into two stages - head[headPos:] and tail.
	// PopFront is trivial (headPos++) on the first stage, and
	// PushBack is trivial (append) on the second stage.
	// If the first stage is Empty, PopFront can swap the
	// first and second stages to remedy the situation.
	//
	// This two-stage split is analogous to the use of two lists
	// in Okasaki's purely functional queue but without the
	// overhead of reversing the list when swapping stages.
	head    []interface{}
	headPos int
	tail    []interface{}
}

// len returns the number of items in the queue.
func (q *Queue) len() int {
	return len(q.head) - q.headPos + len(q.tail)
}

func (q *Queue) Empty() bool {
	return q.len() == 0
}

// PushBack adds w to the back of the queue.
func (q *Queue) PushBack(w interface{}) {
	q.tail = append(q.tail, w)
}

// PopFront removes and returns the element at the front of the queue.
func (q *Queue) PopFront() interface{} {
	if q.headPos >= len(q.head) {
		if len(q.tail) == 0 {
			return nil
		}
		// Pick up tail as new head, clear tail.
		q.head, q.headPos, q.tail = q.tail, 0, q.head[:0]
	}
	w := q.head[q.headPos]
	q.head[q.headPos] = nil
	q.headPos++
	return w
}

// PeekFront returns the P4Folder at the front of the queue without removing it.
func (q *Queue) PeekFront() interface{} {
	if q.headPos < len(q.head) {
		return q.head[q.headPos]
	}
	if len(q.tail) > 0 {
		return q.tail[0]
	}
	return nil
}

// CleanFront pops any P4Folders that are no longer waiting from the head of the
// queue, reporting whether any were popped.
func (q *Queue) CleanFront() (cleaned bool) {
	for {
		w := q.PeekFront()
		if w == nil {
			return cleaned
		}
		q.PopFront()
		cleaned = true
	}
}
