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
	lru     *Lru
	setChan = make(chan interface{}, 1e5)
	Debug   bool
)

type MyList struct {
	*list.List
}

func (l *MyList) PushFront(v *entry) {
	setChan <- v
}

func (l *MyList) MoveToFront(e *list.Element) {
	setChan <- e
}

func worker() {
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("lc:worker() exit unnormal: %v", err)
				}
			}()

			tick := time.Tick(time.Second)
			for {
				select {
				case <-tick:
					if lru.list.Len() > lru.capacity+1000 {
						for lru.list.Len() > lru.capacity {
							e := lru.list.Back()
							lru.list.Remove(e)
							kv := e.Value.(*entry)
							lru.table.Delete(kv.key)
						}
					}

					for i := 1000; i > 0 && lru.list.Len() > 0; i-- {
						e := lru.list.Back()
						kv := e.Value.(*entry)
						if time.Now().Sub(kv.expire) > time.Hour {
							lru.list.Remove(e)
							lru.table.Delete(kv.key)
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
		}()

		time.Sleep(time.Second)
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
		ent = element.(*list.Element).Value.(*entry)
		if ent.ratio++; ent.ratio&0xFF == 0 {
			lru.list.MoveToFront(element.(*list.Element))
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
	if Debug {
		log.Printf("[LC] - [Get] - key: %v\n", key)
	}

	ent, _ok := lru.Get(key)
	if !_ok || ent.key != key {
		ent = Set(key, nil, 0)
	}

	value = ent.value
	if delta := time.Since(ent.expire); delta > 0 {
		if delta > time.Second*15 {
			ent = Set(key, value, 0)
		}
		ok = !atomic.CompareAndSwapInt32(&(ent.expired), 0, 1)
	} else {
		ok = true
	}

	return
}

func Mget(keys []string) (values, valuesAlter map[string]interface{}) {
	if Debug {
		log.Printf("[LC] - [Mget] - keys: %v\n", keys)
	}

	num := len(keys)
	if num == 0 {
		return
	}

	values = make(map[string]interface{}, num)
	valuesAlter = make(map[string]interface{}, num)
	mapFilter := make(map[string]bool, num)

	for _, key := range keys {
		if mapFilter[key] {
			continue
		}
		mapFilter[key] = true
		if value, ok := Get(key); ok {
			values[key] = value
		} else {
			valuesAlter[key] = value
		}
	}
	return
}

func Delete(key string) {
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

// num 限制存储的key的最大个数
func Init(num int) {
	if num <= 0 {
		num = 65536
	}
	lru = NewLru(num)
	go worker()
}
