package cache

import "container/list"

// Cacher is an interface that adds expected methods for memory caching
type Cacher interface {
	Purge()
	Add(key, value interface{}) bool
	Get(key interface{}) (interface{}, bool)
	Remove(key interface{}) bool
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
func NewLRUCache(size int) *LRUCache {
	cache := &LRUCache{
		size:      size,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element),
	}

	return cache
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

	if c.evictList.Len() > c.size && c.size != 0 {
		c.removeOldest()
		return true
	}
	return false
}

// Get returns the value for the given key or nil if it's not found.
func (c *LRUCache) Get(key interface{}) (interface{}, bool) {
	if e, ok := c.items[key]; ok {
		c.evictList.MoveToFront(e)
		return e.Value.(*entry).value, true
	}
	return nil, false
}

// Remove removes the value at the specified key. Returns true if they key was contained
// in the cache.
func (c *LRUCache) Remove(key interface{}) bool {
	if e, ok := c.items[key]; ok {
		c.removeElement(e)
		return true
	}
	return false
}

// Purge removes all entries from the cache.
func (c *LRUCache) Purge() {
	for k := range c.items {
		delete(c.items, k)
	}
	c.evictList.Init()
}

// Len returns the length of the cache.
func (c *LRUCache) Len() int {
	return c.evictList.Len()
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
