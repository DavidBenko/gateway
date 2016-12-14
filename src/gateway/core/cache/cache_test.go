package cache_test

import (
	"gateway/core/cache"
	"testing"

	gc "gopkg.in/check.v1"
)

func TestCache(t *testing.T) { gc.TestingT(t) }

type CacheSuite struct{}

var _ = gc.Suite(&CacheSuite{})

func (s *CacheSuite) TestLRUCacheIsCacher(c *gc.C) {
	lru := cache.NewLRUCache(5)

	if _, ok := interface{}(lru).(cache.Cacher); !ok {
		c.Error("LRUCache does not implement Cacher interface")
	}
}

func (s *CacheSuite) TestNewLRUCache(c *gc.C) {
	for i, t := range []struct {
		should string
		size   int
	}{{
		should: "create a cache with a valid size",
		size:   5,
	}, {
		should: "create a cache with size 0",
		size:   0,
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		cache := cache.NewLRUCache(t.size)
		c.Assert(cache, gc.NotNil)
	}
}

func (s *CacheSuite) TestLRUCacheContains(c *gc.C) {
	cache := cache.NewLRUCache(5)

	// Should not contains "foo"
	ok := cache.Contains("foo")
	c.Assert(ok, gc.Equals, false)

	// Add the key, value. Should not evict anything from the cache
	evicted := cache.Add("foo", "bar")
	c.Assert(evicted, gc.Equals, false)

	// Cache should contain "foo" now
	ok = cache.Contains("foo")
	c.Assert(ok, gc.Equals, true)
}

func (s *CacheSuite) TestLRUCacheAdd(c *gc.C) {
	cache := cache.NewLRUCache(1)

	c.Assert(cache.Len(), gc.Equals, 0)

	// Cache should contain struct and Len() should 1 after adding a value
	cache.Add("foo", "bar")
	c.Assert(cache.Contains("foo"), gc.Equals, true)
	c.Assert(cache.Len(), gc.Equals, 1)

	// Add a value for an existing key should not change the Len()
	cache.Add("foo", "baz")
	c.Assert(cache.Contains("foo"), gc.Equals, true)
	c.Assert(cache.Len(), gc.Equals, 1)

	// Adding a new key should cause an eviction since the cache size is 1
	cache.Add("bar", "baz")
	c.Assert(cache.Contains("bar"), gc.Equals, true)
	// No longer contains foo
	c.Assert(cache.Contains("foo"), gc.Equals, false)
	c.Assert(cache.Len(), gc.Equals, 1)
}

func (s *CacheSuite) TestLRUCacheUnlimitedSize(c *gc.C) {
	cache := cache.NewLRUCache(0)
	c.Assert(cache.Len(), gc.Equals, 0)

	cache.Add("foo", "bar")
	c.Assert(cache.Len(), gc.Equals, 1)
}

func (s *CacheSuite) TestLRUCacheGet(c *gc.C) {
	cache := cache.NewLRUCache(1)

	c.Assert(cache.Len(), gc.Equals, 0)

	cache.Add("foo", "bar")

	v, ok := cache.Get("foo")
	c.Assert(ok, gc.Equals, true)
	c.Assert(v, gc.Equals, "bar")

	v, ok = cache.Get("invalid")
	c.Assert(ok, gc.Equals, false)
	c.Assert(v, gc.IsNil)
}

func (s *CacheSuite) TestLRUCachePurge(c *gc.C) {
	cache := cache.NewLRUCache(5)

	c.Assert(cache.Len(), gc.Equals, 0)

	keys := []string{"a", "b", "c", "d", "e"}
	for v := range keys {
		cache.Add(v, &struct{}{})
	}

	c.Assert(cache.Len(), gc.Equals, 5)
	cache.Purge()
	c.Assert(cache.Len(), gc.Equals, 0)
}
