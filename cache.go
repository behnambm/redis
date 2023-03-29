package main

import (
	"errors"
	"sync"
	"time"
)

var (
	ValueExpiredError = errors.New("value expired")
	NotFoundError     = errors.New("not found")
)

type CacheValue struct {
	value      string
	expiration time.Time
}

type MemoryCache struct {
	mu   sync.RWMutex
	data map[string]CacheValue
}

func New() *MemoryCache {
	return &MemoryCache{data: map[string]CacheValue{}}
}

func (c *MemoryCache) Set(key string, value CacheValue) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
}

func (c *MemoryCache) Get(key string) (CacheValue, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.data[key]
	if !ok {
		return CacheValue{}, NotFoundError
	}
	if value.IsExpired() {
		return CacheValue{}, ValueExpiredError
	}
	return value, nil

}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

func (v *CacheValue) IsExpired() bool {
	if v.expiration.IsZero() {
		return false
	}

	return v.expiration.Before(time.Now())
}
