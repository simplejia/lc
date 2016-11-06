package lc

import (
	"time"

	"github.com/simplejia/utils"
)

type Elem struct {
	Key   string
	Value interface{}
	Birth int64
}

type HashMap struct {
	elems []*Elem
	bnum  int
	blen  int
}

func (this *HashMap) Init(num int) *HashMap {
	if this == nil {
		this = new(HashMap)
	}
	this.blen = 100
	capacity := num / this.blen * this.blen
	this.bnum = capacity / this.blen
	this.elems = make([]*Elem, capacity, capacity)

	return this
}

func (this *HashMap) getElem(key string) (elem *Elem, pos int) {
	hash := utils.Hash33(key)
	index := (hash % this.bnum) * this.blen

	var oldRecord *Elem
	var freeRecordFlag bool
	for i := 0; i < this.blen; i++ {
		j := index + i
		tmpRecord := this.elems[j]
		if tmpRecord == nil {
			pos = j
			freeRecordFlag = true
			continue
		}
		if tmpRecord.Key == key {
			pos = j
			elem = tmpRecord
			break
		}
		if freeRecordFlag {
			continue
		}
		if oldRecord == nil || oldRecord.Birth > tmpRecord.Birth {
			oldRecord = tmpRecord
			pos = j
			continue
		}
	}

	return
}

func (this *HashMap) Len() int {
	return len(this.elems)
}

func (this *HashMap) Get(key string) (value interface{}, ok bool) {
	elem, _ := this.getElem(key)
	if elem != nil {
		value, ok = elem.Value, true
	}
	return
}

func (this *HashMap) Set(key string, value interface{}) {
	now := time.Now().Unix()
	elem, pos := this.getElem(key)
	if elem != nil {
		elem.Value, elem.Birth = value, now
	} else {
		this.elems[pos] = &Elem{
			Key:   key,
			Value: value,
			Birth: now,
		}
	}
	return
}

func (this *HashMap) Delete(key string) {
	elem, pos := this.getElem(key)
	if elem != nil {
		this.elems[pos] = nil
	}
	return
}
