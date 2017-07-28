[中文 README](#中文)


# [lc](http://github.com/simplejia/lc) (local cache)
## Original Intention
* Compared with lc, only using redis to cache is with these disadvantages such as network overhead, latency increasing dramatically if call repeatedly. In addition, using redis, when there is something wrong with the network, we cannot get data from redis. But in lc, even if the data is expired, we still can get the expired data. 
* When the cache is invalid, using mysql will take the risk of data penetration. While lc is with concurrency control which allows only one client of the same key at the same time to penetrate to the database. All other clients will return to the lc cache data.

## Features
* Local cache
* Supporting the operation of Get，Set，Mget，Delete
* When the cache is invalid, it not only returns the invalid flag, but also returns old data, for example: v, ok := lc.Get(key). When key expired and has not been deleted by lru, v is the stored value and ok returns false.
* Lock free
* Lru algorithm
* Combining with [lm](http://github.com/simplejia/lm), much simpler and quicker to use

## demo
[lc_test.go](http://github.com/simplejia/lc/tree/master/lc_test.go)
```
package lc

import (
	"testing"
	"time"
)

func init() {
	Init(65536) // must init
}

func TestGetValid(t *testing.T) {
	key := "k"
	value := "v"
	Set(key, value, time.Second)
	time.Sleep(time.Millisecond * 10) // wait a while
	v, ok := Get(key)
	if !ok || v != value {
		t.Fatal("")
	}
}
```

---
中文
===

# [lc](http://github.com/simplejia/lc) (local cache)
## 实现初衷
* 纯用redis做缓存，相比lc，redis有网络调用开销，反复调用多次，延时急剧增大，当网络偶尔出现故障时，我们的数据接口也就拿不到数据，但lc里的数据就算是超过了设置的过期时间，我们一样能拿到过期的数据做备用 
* 使用mysql，当缓存失效，有数据穿透的风险，lc自带并发控制，有且只允许同一时间同一个key的唯一一个client穿透到数据库，其它直接返回lc缓存数据

## 特性
* 本地缓存
* 支持Get，Set，Mget，Delete操作
* 当缓存失效时，返回失效标志同时，还返回旧的数据，如：v, ok := lc.Get(key)，当key已经过了失效时间了，并且key还没有被lru淘汰掉，v是之前存的值，ok返回false
* 实现代码没有用到锁
* 使用到lru，淘汰长期不用的key
* 结合[lm](http://github.com/simplejia/lm)使用更简单快捷

## demo

一般使用方式：
```
vLc, ok := lc.Get(key)
v, _ := vLc.(int)
if ok {
   return
}
```

[lc_test.go](http://github.com/simplejia/lc/tree/master/lc_test.go)
```
package lc

import (
	"testing"
	"time"
)

func init() {
	Init(65536) // 使用lc之前必须要初始化
}

func TestGetValid(t *testing.T) {
	key := "k"
	value := "v"
	Set(key, value, time.Second)
	time.Sleep(time.Millisecond * 10) // 给异步处理留点时间
	v, ok := Get(key)
	if !ok || v != value {
		t.Fatal("")
	}
}
```
