package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	if listItem, ok := c.items[key]; ok {
		listItem.Value = cacheItem{key, value}
		c.queue.MoveToFront(listItem)
		return ok
	}

	if c.capacity == c.queue.Len() {
		listItem := c.queue.Back()
		cacheElement := listItem.Value.(cacheItem)
		delete(c.items, cacheElement.key)
		c.queue.Remove(listItem)
	}

	c.items[key] = c.queue.PushFront(cacheItem{key, value})
	return false
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	if listItem, ok := c.items[key]; ok {
		c.queue.MoveToFront(listItem)
		cacheElement := listItem.Value.(cacheItem)
		return cacheElement.value, true
	}
	return nil, false
}

func (c *lruCache) Clear() {
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}

type cacheItem struct {
	key   Key
	value interface{}
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
