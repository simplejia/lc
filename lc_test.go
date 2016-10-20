package lc

import (
	"testing"
	"time"
)

func init() {
	Init(65536)
}

func set(key string, value interface{}, expire time.Duration) {
	Set(key, value, expire)
	time.Sleep(time.Millisecond * 10) // 给异步处理留点时间
}

func TestGetValid(t *testing.T) {
	key := "k"
	value := "v"
	set(key, value, time.Second)
	v, ok := Get(key)
	if !ok || v != value {
		t.Fatal("fail")
	}
}

func TestGetInvalid(t *testing.T) {
	key := "k"
	value := "v"
	set(key, value, time.Millisecond*5)
	time.Sleep(time.Millisecond * 10)
	v, ok := Get(key)
	if ok || v != value {
		t.Fatal("fail")
	}
}

func TestMgetValid(t *testing.T) {
	key := "k"
	value := "v"
	set(key, value, time.Second)
	key1 := "k1"
	value1 := "v1"
	set(key1, value1, time.Second)

	vs, _ := Mget([]string{key, key1})
	if vs[key] != value || vs[key1] != value1 {
		t.Fatal("fail")
	}
}

func TestMgetInvalid(t *testing.T) {
	key := "k"
	value := "v"
	set(key, value, time.Second)
	key1 := "k1"
	value1 := "v1"
	set(key1, value1, time.Millisecond*5)
	time.Sleep(time.Millisecond * 10)

	vs, vsAlter := Mget([]string{key, key1})
	if vs[key] != value || vsAlter[key1] != value1 {
		t.Fatal("fail")
	}
}
