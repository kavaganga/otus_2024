package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	mu       sync.Mutex
	capacity int
	queue    List
	items    map[Key]*ListItem
	keys     map[*ListItem]Key
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
		keys:     make(map[*ListItem]Key, capacity),
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	c.mu.Lock()
	item, ok := c.items[key]

	if ok {
		item.Value = value
		c.queue.MoveToFront(item)
	} else {
		item = c.queue.PushFront(value)
		c.items[key] = item
		c.keys[item] = key

		length := c.queue.Len()
		if c.capacity > 0 && length > c.capacity {
			back := c.queue.Back()
			backKey := c.keys[back]
			delete(c.items, backKey)
			delete(c.keys, back)
			c.queue.Remove(c.queue.Back())
		}
	}
	c.mu.Unlock()
	return ok
}

func (c *lruCache) Clear() {
	c.mu.Lock()
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
	c.mu.Unlock()
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if ok {
		c.queue.MoveToFront(item)

		return item.Value, ok
	}
	return nil, false
}
