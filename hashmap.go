// NOTICE: hashmap is not ok when enable data race detection.

package lc

import (
	"time"

	"github.com/simplejia/utils"
)

type Elem struct {
	key   string
	value interface{}
	birth int64
}

type HashMap struct {
	elems []*Elem
	bnum  int
	blen  int
}

func (m *HashMap) Init(num int) *HashMap {
	if m == nil {
		m = new(HashMap)
	}

	m.blen = 100
	m.bnum = int(float64(num)*1.2)/m.blen + 1
	m.elems = make([]*Elem, m.blen*m.bnum)

	return m
}

func (m *HashMap) getElem(key string) (elem *Elem, pos int) {
	hash := utils.Hash33(key)
	index := (hash % m.bnum) * m.blen

	var oldRecord *Elem
	var freeRecordFlag bool
	for i := 0; i < m.blen; i++ {
		j := index + i
		tmpRecord := m.elems[j]
		if tmpRecord == nil {
			pos = j
			freeRecordFlag = true
			continue
		}
		if tmpRecord.key == key {
			pos = j
			elem = tmpRecord
			break
		}
		if freeRecordFlag {
			continue
		}
		if oldRecord == nil || oldRecord.birth > tmpRecord.birth {
			oldRecord = tmpRecord
			pos = j
			continue
		}
	}

	return
}

func (m *HashMap) Len() int {
	return len(m.elems)
}

func (m *HashMap) Get(key string) (value interface{}, ok bool) {
	elem, _ := m.getElem(key)
	if elem != nil {
		value, ok = elem.value, true
	}
	return
}

func (m *HashMap) Set(key string, value interface{}) {
	now := time.Now().Unix()
	elem, pos := m.getElem(key)
	if elem != nil {
		elem.value, elem.birth = value, now
	} else {
		m.elems[pos] = &Elem{
			key:   key,
			value: value,
			birth: now,
		}
	}
	return
}

func (m *HashMap) Delete(key string) {
	elem, pos := m.getElem(key)
	if elem != nil {
		m.elems[pos] = nil
	}
	return
}
