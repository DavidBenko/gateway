package cache

import (
	"container/list"
	"errors"
)

// Cacher is an interface that adds expected methods for memory caching
type Cacher interface {
	Purge()
	Add(key, value interface{}) bool
	Get(key interface{}) (interface{}, bool)
	Contains(key interface{}) bool
}

type entry struct {
	key   interface{}
	value interface{}
}

// LRUCache is a simple LRU cache implementation.
type LRUCache struct {
	size      int
	items     map[interface{}]*list.Element
	evictList *list.List
}

// NewLRUCache returns a new LRU cache of the given size.
func NewLRUCache(size int) (*LRUCache, error) {
	if size <= 0 {
		return nil, errors.New("size must be greater than 0")
	}

	cache := &LRUCache{
		size:      size,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element),
	}

	return cache, nil
}

// Contains checks if a given key exists in the cache.
func (c *LRUCache) Contains(key interface{}) bool {
	_, ok := c.items[key]
	return ok
}

// Add adds the supplied value to the cache at the given key. Returns true if
// an eviction occurred.
func (c *LRUCache) Add(key, value interface{}) bool {
	if e, ok := c.items[key]; ok {
		c.evictList.MoveToFront(e)
		e.Value.(*entry).value = value
		return false
	}

	e := &entry{key, value}
	element := c.evictList.PushFront(e)
	c.items[key] = element

	if c.evictList.Len() > c.size {
		c.removeOldest()
		return true
	}
	return false
}

func (c *LRUCache) Get(key interface{}) (interface{}, bool) {
	if e, ok := c.items[key]; ok {
		c.evictList.MoveToFront(e)
		return e.Value.(*entry).value, true
	}
	return nil, false
}

func (c *LRUCache) Purge() {
	for k, _ := range c.items {
		delete(c.items, k)
	}
	c.evictList.Init()
}

func (c *LRUCache) removeOldest() {
	entry := c.evictList.Back()
	if entry != nil {
		c.removeElement(entry)
	}
}

func (c *LRUCache) removeElement(element *list.Element) {
	c.evictList.Remove(element)
	e := element.Value.(*entry)
	delete(c.items, e.key)
}

func (c *LRUCache) Len() int {
	return c.evictList.Len()
}
