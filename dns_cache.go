package main

import (
	"sync"
	"time"
	"fmt"
)

type DnsCache struct {
	mu sync.RWMutex
	m  map[string]*CacheItem
}

type CacheItem struct {
	Record *DnsRecord
	Expiry time.Time
}

func NewDnsCache() *DnsCache {
	return &DnsCache{
		m: make(map[string]*CacheItem),
	}
}

func cacheKey(name string, qtype QType) string {
	return fmt.Sprintf("%s|%d", name, uint16(qtype))
}

func (c *DnsCache) Get(name string, qtype QType) (*DnsRecord, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	it, ok := c.m[cacheKey(name, qtype)]
	if !ok {
		return nil, false
	}
	if time.Now().After(it.Expiry) {
		// expired
		return nil, false
	}
	return it.Record, true
}

func (c *DnsCache) Put(name string, qtype QType, rec *DnsRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	exp := time.Now().Add(time.Duration(rec.TTL) * time.Second)
	c.m[cacheKey(name, qtype)] = &CacheItem{Record: rec, Expiry: exp}
}

func (c *DnsCache) PutMultiple(records []*DnsRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, r := range records {
		key := cacheKey(r.Name, r.Type)
		c.m[key] = &CacheItem{Record: r, Expiry: time.Now().Add(time.Duration(r.TTL) * time.Second)}
	}
}

func (c *DnsCache) Stats() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	total := len(c.m)
	active := 0
	for k, v := range c.m {
		if time.Now().Before(v.Expiry) && v.Record != nil {
			active++
		} else {
			delete(c.m, k)
		}
	}
	return fmt.Sprintf("Total: %d, Active: %d", total, active)
}

func (c *DnsCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, v := range c.m {
		if now.After(v.Expiry) {
			delete(c.m, k)
		}
	}
}
