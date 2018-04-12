// Package lc is A local cache for golang.
// Created by simplejia [7/2015]
package lc

import (
	"container/list"
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

var (
	lru      *Lru
	setChan  = make(chan interface{}, 1e5)
	Debug    bool
	Disabled bool
)

type MyList struct {
	*list.List
}

func (l *MyList) PushFront(v interface{}) *list.Element {
	setChan <- v
	return nil
}

func (l *MyList) MoveToFront(e *list.Element) {
	setChan <- e
}

func worker() {
	tick := time.Tick(time.Second)
	for {
		select {
		case <-tick:
			if lru.list.Len() > lru.capacity+100 {
				for lru.list.Len() > lru.capacity {
					e := lru.list.Back()
					lru.list.Remove(e)
					ent := e.Value.(*entry)
					lru.table.Delete(ent.key)
				}
			}

			for i := 100; i > 0 && lru.list.Len() > 0; i-- {
				e := lru.list.Back()
				ent := e.Value.(*entry)
				if time.Since(ent.expire) > time.Hour {
					lru.list.Remove(e)
					lru.table.Delete(ent.key)
				} else {
					break
				}
			}
		case itm := <-setChan:
			switch v := itm.(type) {
			case *list.Element:
				lru.list.List.MoveToFront(v)
			case *entry:
				e := lru.list.List.PushFront(v)
				lru.table.Set(v.key, e)
			}
		}
	}
}

type Lru struct {
	list     *MyList
	table    *HashMap
	capacity int
}

type entry struct {
	key     string
	value   interface{}
	expired int32
	expire  time.Time
	ratio   uint32
}

func NewLru(capacity int) *Lru {
	return &Lru{
		list:     &MyList{list.New()},
		table:    (&HashMap{}).Init(capacity),
		capacity: capacity,
	}
}

func (lru *Lru) Get(key string) (ent *entry, ok bool) {
	element, ok := lru.table.Get(key)
	if ok {
		e := element.(*list.Element)
		ent = e.Value.(*entry)
		if ent.ratio++; ent.ratio&0xFF == 0 {
			lru.list.MoveToFront(e)
		}
	}
	return
}

func (lru *Lru) Set(ent *entry) {
	if element, ok := lru.table.Get(ent.key); ok {
		element.(*list.Element).Value = ent
	} else {
		lru.list.PushFront(ent)
	}
}

func (lru *Lru) Delete(key string) {
	if element, ok := lru.table.Get(key); ok {
		ent := element.(*list.Element).Value.(*entry)
		ent.expire = time.Now()
	}
}

func Set(key string, value interface{}, expire time.Duration) (ent *entry) {
	if Disabled {
		return
	}
	if Debug {
		log.Printf("[LC] - [Set] - key: %v, value: %v, expire: %v\n", key, value, expire)
	}
	ent = &entry{
		key:    key,
		value:  value,
		expire: time.Now().Add(expire),
	}
	lru.Set(ent)
	return
}

func Get(key string) (value interface{}, ok bool) {
	if Disabled {
		return
	}
	if Debug {
		log.Printf("[LC] - [Get] - key: %v\n", key)
	}

	ent, _ok := lru.Get(key)
	if !_ok || ent.key != key {
		ent = Set(key, nil, -1)
	}

	value = ent.value
	if delta := time.Since(ent.expire); delta > 0 {
		if delta > time.Second*10 {
			ent = Set(key, value, 0)
		}
		ok = !atomic.CompareAndSwapInt32(&(ent.expired), 0, 1)
	} else {
		ok = true
	}

	return
}

func Mget(keys []string) (values, valuesAlter map[string]interface{}) {
	if Disabled {
		return
	}
	if Debug {
		log.Printf("[LC] - [Mget] - keys: %v\n", keys)
	}

	num := len(keys)
	if num == 0 {
		return
	}

	values = map[string]interface{}{}
	valuesAlter = map[string]interface{}{}

	for _, key := range keys {
		if value, ok := Get(key); ok {
			values[key] = value
		} else {
			valuesAlter[key] = value
		}
	}
	return
}

func Delete(key string) {
	if Disabled {
		return
	}
	if Debug {
		log.Printf("[LC] - [Delete] - key: %v\n", key)
	}
	lru.Delete(key)
}

func GetAll() (items []string) {
	for e := lru.list.Front(); e != nil; e = e.Next() {
		ent := e.Value.(*entry)
		item := fmt.Sprintf("%v, %v, %v, %v, %v",
			ent.key, ent.value, ent.expired, ent.expire, ent.ratio)
		items = append(items, item)
	}
	return
}

func Init(num int) {
	if num <= 0 {
		num = 65536
	}
	lru = NewLru(num)
	go worker()
}
