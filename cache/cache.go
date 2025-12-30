package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// Cache represents an in-memory cache with TTL support
type Cache struct {
	data      map[string]*CacheEntry
	mu        sync.RWMutex
	maxSize   int
	defaultTTL time.Duration
}

// CacheEntry represents a single cache entry with TTL
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size       int
	Hits       int64
	Misses     int64
	Evictions  int64
	HitRate    float64
}

// NewCache creates a new cache instance
func NewCache(maxSize int, defaultTTL time.Duration) *Cache {
	c := &Cache{
		data:       make(map[string]*CacheEntry),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
	}

	// Start cleanup routine for expired entries
	go c.cleanupExpired()

	return c
}

// Set sets a value in the cache with default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL sets a value in the cache with custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple eviction when cache is full: remove oldest entry
	if len(c.data) >= c.maxSize && c.data[key] == nil {
		c.evictOldest()
	}

	c.data[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.data[key]
	if !exists {
		c.mu.RUnlock()
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		return nil, false
	}

	val := entry.Value
	c.mu.RUnlock()
	return val, true
}

// GetJSON retrieves and unmarshals a JSON value from cache
func (c *Cache) GetJSON(key string, v interface{}) bool {
	val, exists := c.Get(key)
	if !exists {
		return false
	}

	// Try to unmarshal if it's JSON bytes
	if jsonBytes, ok := val.([]byte); ok {
		return json.Unmarshal(jsonBytes, v) == nil
	}

	// Otherwise try to cast
	if jsonStr, ok := val.(string); ok {
		return json.Unmarshal([]byte(jsonStr), v) == nil
	}

	return false
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Clear removes all values from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]*CacheEntry)
}

// Exists checks if a key exists and is not expired
func (c *Cache) Exists(key string) bool {
	_, exists := c.Get(key)
	return exists
}

// Size returns the current number of entries in cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// evictOldest removes the oldest entry from cache
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.data {
		if oldestTime.IsZero() || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(c.data, oldestKey)
	}
}

// cleanupExpired periodically removes expired entries
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.ExpiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// ============================================================================
// Specialized Cache for Leave Balances
// ============================================================================

type LeaveBalanceCache struct {
	cache *Cache
}

// NewLeaveBalanceCache creates a cache for leave balances (1 hour TTL, 1000 entries)
func NewLeaveBalanceCache() *LeaveBalanceCache {
	return &LeaveBalanceCache{
		cache: NewCache(1000, 1*time.Hour),
	}
}

// SetLeaveBalance caches leave balance for an employee
func (lc *LeaveBalanceCache) SetLeaveBalance(ctx context.Context, employeeID int64, balance map[string]interface{}) {
	key := generateLeaveBalanceKey(employeeID)
	data, _ := json.Marshal(balance)
	lc.cache.SetWithTTL(key, data, 1*time.Hour)
}

// GetLeaveBalance retrieves cached leave balance
func (lc *LeaveBalanceCache) GetLeaveBalance(ctx context.Context, employeeID int64) (map[string]interface{}, bool) {
	key := generateLeaveBalanceKey(employeeID)
	var balance map[string]interface{}
	exists := lc.cache.GetJSON(key, &balance)
	return balance, exists
}

// InvalidateLeaveBalance removes cached balance for an employee
func (lc *LeaveBalanceCache) InvalidateLeaveBalance(ctx context.Context, employeeID int64) {
	key := generateLeaveBalanceKey(employeeID)
	lc.cache.Delete(key)
}

// ============================================================================
// Specialized Cache for Employee Data
// ============================================================================

type EmployeeCache struct {
	cache *Cache
}

// NewEmployeeCache creates a cache for employee data (2 hour TTL, 500 entries)
func NewEmployeeCache() *EmployeeCache {
	return &EmployeeCache{
		cache: NewCache(500, 2*time.Hour),
	}
}

// SetEmployee caches employee data
func (ec *EmployeeCache) SetEmployee(ctx context.Context, employeeID int64, data interface{}) {
	key := generateEmployeeKey(employeeID)
	jsonData, _ := json.Marshal(data)
	ec.cache.SetWithTTL(key, jsonData, 2*time.Hour)
}

// GetEmployee retrieves cached employee data
func (ec *EmployeeCache) GetEmployee(ctx context.Context, employeeID int64, v interface{}) bool {
	key := generateEmployeeKey(employeeID)
	return ec.cache.GetJSON(key, v)
}

// InvalidateEmployee removes cached employee data
func (ec *EmployeeCache) InvalidateEmployee(ctx context.Context, employeeID int64) {
	key := generateEmployeeKey(employeeID)
	ec.cache.Delete(key)
}

// ============================================================================
// Helper Functions
// ============================================================================

func generateLeaveBalanceKey(employeeID int64) string {
	return "leave_balance:" + string(rune(employeeID))
}

func generateEmployeeKey(employeeID int64) string {
	return "employee:" + string(rune(employeeID))
}

// InvalidateAllCaches clears all caches - call this on critical updates
func InvalidateAllCaches(leaveCache *LeaveBalanceCache, empCache *EmployeeCache) {
	if leaveCache != nil {
		leaveCache.cache.Clear()
	}
	if empCache != nil {
		empCache.cache.Clear()
	}
}
