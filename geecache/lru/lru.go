package lru

import "container/list"

// 双端队列 + Map  实现LRU Cache
type Cache struct {
	maxBytes  int64
	nBytes    int64
	ll        *list.List
	cache     map[string]*list.Element   // map中保存的不是value，而是经过封装的Element
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}


// 构造函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// ------------------  功能函数 --------------------------

// 新增/修改
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { //该值已经存在
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry) // 获取map中该entry的地址
		c.nBytes -= int64(value.Len()) - int64(kv.value.Len())
		kv.value = value    // 修改map中保存的值
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele   // map中保存的不是value，而是经过封装的*list.Element
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	// 清除最久未使用，知道当前缓存小于最大缓存的限制
	for c.maxBytes != 0 && c.nBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

// 查找
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)    //移动到队首 （双向链表队首队尾是相对的，在这里约定 front 为队首）
		kv := ele.Value.(*entry) // TODO: 这一句是什么意思？得到map中该位置的指针？？
		return kv.value, true
	}
	return
}


// 删除
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
	return
}

// 链表的长度
func (c *Cache) Len() int{
	return c.ll.Len()
}

