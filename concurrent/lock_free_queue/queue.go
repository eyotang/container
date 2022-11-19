/*
 *
 * queue - Goroutine-safe LockFreeQueue implementations
 * Copyright (C) 2016 Antigloss Huang (https://github.com/antigloss) All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

// Package queue offers goroutine-safe LockFreeQueue implementations such as LockfreeQueue(Lock free queue).
package lock_free_queue

import (
	"sync/atomic"
	"unsafe"
)

// LockFreeQueue is a goroutine-safe LockFreeQueue implementation.
// The overall performance of LockFreeQueue is much better than List+Mutex(standard package).
type LockFreeQueue[T any] struct {
	head  unsafe.Pointer
	tail  unsafe.Pointer
	dummy qNode[T]
}

// NewQueue is the only way to get a new, ready-to-use LockfreeQueue.
//
// Example:
//
//	lfq := queue.NewQueue[int]()
//	lfq.Push(100)
//	v, ok := lfq.Pop()
func NewQueue[T any]() *LockFreeQueue[T] {
	var queue LockFreeQueue[T]
	queue.head = unsafe.Pointer(&queue.dummy)
	queue.tail = queue.head
	return &queue
}

// Pop returns (and removes) an element from the front of the queue and true if the queue is not empty,
// otherwise it returns a default value and false if the queue is empty.
// It performs about 100% better than list.List.Front() and list.List.Pop() with sync.Mutex.
func (queue *LockFreeQueue[T]) Pop() (T, bool) {
	for {
		h := atomic.LoadPointer(&queue.head)
		rh := (*qNode[T])(h)
		n := (*qNode[T])(atomic.LoadPointer(&rh.next))
		if n != nil {
			if atomic.CompareAndSwapPointer(&queue.head, h, rh.next) {
				return n.val, true
			} else {
				continue
			}
		} else {
			var v T
			return v, false
		}
	}
}

// Push inserts an element to the back of the queue.
// It performs exactly the same as list.List.PushBack() with sync.Mutex.
func (queue *LockFreeQueue[T]) Push(val T) {
	node := unsafe.Pointer(&qNode[T]{val: val})
	for {
		rt := (*qNode[T])(atomic.LoadPointer(&queue.tail))
		//t := atomic.LoadPointer(&queue.tail)
		//rt := (*qNode[T])(t)
		if atomic.CompareAndSwapPointer(&rt.next, nil, node) {
			atomic.StorePointer(&queue.tail, node)
			// If dead loop occurs, use CompareAndSwapPointer instead of StorePointer
			// atomic.CompareAndSwapPointer(&queue.tail, t, node)
			return
		} else {
			continue
		}
	}
}

type qNode[T any] struct {
	val  T
	next unsafe.Pointer
}
