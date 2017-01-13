package gotools

import (
	"container/heap"
	"time"
)

type pair struct {
	key interface{}
	val interface{}
}

type node struct {
	ts   time.Time
	pos  int // pos in []node
	data pair
}

type myHeap []*node

func (h *myHeap) Less(i, j int) bool {
	return (*h)[i].ts.Before((*h)[j].ts)
}

func (h *myHeap) Swap(i, j int) {
	(*h)[i].pos, (*h)[j].pos = (*h)[j].pos, (*h)[i].pos
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *myHeap) Len() int {
	return len(*h)
}

func (h *myHeap) Push(v interface{}) {
	*h = append(*h, v.(*node))
}

func (h *myHeap) Pop() (v interface{}) {
	*h, v = (*h)[:h.Len()-1], (*h)[h.Len()-1]
	return
}

type ExpiredMap struct {
	m map[interface{}]*node
	h *myHeap
	n int
}

const defaultLen = 4096

func NewExpiredMap(limit int) *ExpiredMap {
	if limit <= 0 {
		panic("Limitation of Expired Map CANNOT less than or equal to ZERO")
	}
	em := &ExpiredMap{
		m: make(map[interface{}]*node),
		n: limit,
	}
	hp := make(myHeap, 0, defaultLen)
	em.h = &hp
	heap.Init(em.h)
	return em
}

func (em *ExpiredMap) remove(index int) {
	if index < em.h.Len() {
		// Remove item from map
		delete(em.m, (*em.h)[index].data.key)

		// Remove item from heap
		heap.Remove(em.h, index)
	}
}

func (em *ExpiredMap) Update() {
	for em.h.Len() > 0 && (*em.h)[0].ts.Before(time.Now()) {
		em.remove(0)
	}
}

func (em *ExpiredMap) Len() int {
	return len(em.m)
}

func (em *ExpiredMap) Del(key interface{}) {
	em.Update()
	n := em.m[key]
	if n != nil {
		em.remove(n.pos)
	}
}

func (em *ExpiredMap) Put(key interface{}, val interface{}, exp time.Duration) {
	em.Update()
	for len(em.m) >= em.n {
		// Remove last the elem to be expired
		em.remove(0)
	}
	if n := em.m[key]; n != nil {
		// the item has been existed, do update
		n.data.val = val
		n.ts = time.Now().Add(exp)
		heap.Fix(em.h, n.pos)
	} else {
		// insert
		n = &node{
			data: pair{key: key, val: val},
			pos:  em.h.Len(),
			ts:   time.Now().Add(exp),
		}
		heap.Push(em.h, n)
		em.m[key] = n
	}
}

func (em *ExpiredMap) Get(key interface{}) interface{} {
	em.Update()
	if n := em.m[key]; n != nil {
		return n.data.val
	}
	return nil
}

func (em *ExpiredMap) NewIterator() *ExpiredMapIter {
	return NewExpiredMap(em)
}

type ExpiredMapIter struct {
	em    *ExpiredMap
	index int
}

func NewExpiredMapIter(em *ExpiredMap) *ExpiredMapIter {
	return &ExpiredMapIter{
		em:    em,
		index: 0,
	}
}

func (iter *ExpiredMapIter) Valid() bool {
	return iter.index >= 0 && iter.index < iter.em.h.Len()
}

func (iter *ExpiredMapIter) Next() {
	if iter.Valid() {
		iter.index++
	}
}

func (iter *ExpiredMapIter) Prev() {
	if iter.Valid() {
		iter.index--
	}
}

func (iter *ExpiredMapIter) Key() interface{} {
	return (*iter.em.h)[iter.index].data.key
}

func (iter *ExpiredMapIter) Val() interface{} {
	return (*iter.em.h)[iter.index].data.val
}

func (iter *ExpiredMapIter) Ts() time.Time {
	assert((*iter.em.h)[iter.index].pos == iter.index)
	return (*iter.em.h)[iter.index].ts
}
