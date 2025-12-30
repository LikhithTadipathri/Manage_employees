package cache_test

import (
	"context"
	"testing"
	"time"

	"employee-service/cache"
)

// TestCacheBasicOperations tests basic cache Get/Set operations
func TestCacheBasicOperations(t *testing.T) {
	c := cache.NewCache(100, 1*time.Hour)

	t.Run("Set and Get value", func(t *testing.T) {
		c.Set("key1", "value1")
		val, exists := c.Get("key1")

		if !exists {
			t.Error("Expected key to exist")
		}

		if val != "value1" {
			t.Errorf("Expected 'value1', got %v", val)
		}
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		_, exists := c.Get("nonexistent")
		if exists {
			t.Error("Expected key to not exist")
		}
	})

	t.Run("Delete key", func(t *testing.T) {
		c.Set("delkey", "delvalue")
		c.Delete("delkey")
		_, exists := c.Get("delkey")

		if exists {
			t.Error("Expected key to be deleted")
		}
	})

	t.Run("Clear all cache", func(t *testing.T) {
		c.Set("key2", "value2")
		c.Set("key3", "value3")
		c.Clear()

		if c.Size() != 0 {
			t.Errorf("Expected cache size 0, got %d", c.Size())
		}
	})
}

// TestCacheTTL tests cache TTL expiration
func TestCacheTTL(t *testing.T) {
	c := cache.NewCache(100, 1*time.Hour)

	t.Run("SetWithTTL and expiration", func(t *testing.T) {
		c.SetWithTTL("short", "value", 100*time.Millisecond)
		val, exists := c.Get("short")

		if !exists {
			t.Error("Expected key to exist immediately after set")
		}

		if val != "value" {
			t.Errorf("Expected 'value', got %v", val)
		}

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		_, exists = c.Get("short")
		if exists {
			t.Error("Expected key to be expired")
		}
	})
}

// TestLeaveBalanceCache tests specialized leave balance caching
func TestLeaveBalanceCache(t *testing.T) {
	lc := cache.NewLeaveBalanceCache()
	ctx := context.Background()

	t.Run("Set and get leave balance", func(t *testing.T) {
		balance := map[string]interface{}{
			"sick_leave":     5,
			"casual_leave":   3,
			"maternity_leave": 180,
		}

		lc.SetLeaveBalance(ctx, 1, balance)
		retrieved, exists := lc.GetLeaveBalance(ctx, 1)

		if !exists {
			t.Error("Expected leave balance to exist")
		}

		if val, ok := retrieved["sick_leave"]; !ok || val != float64(5) {
			t.Errorf("Expected sick_leave to be 5, got %v", retrieved["sick_leave"])
		}
	})

	t.Run("Invalidate leave balance", func(t *testing.T) {
		balance := map[string]interface{}{"sick_leave": 10}
		lc.SetLeaveBalance(ctx, 2, balance)

		lc.InvalidateLeaveBalance(ctx, 2)
		_, exists := lc.GetLeaveBalance(ctx, 2)

		if exists {
			t.Error("Expected balance to be invalidated")
		}
	})
}

// TestEmployeeCache tests specialized employee data caching
func TestEmployeeCache(t *testing.T) {
	ec := cache.NewEmployeeCache()
	ctx := context.Background()

	t.Run("Set and get employee", func(t *testing.T) {
		employee := map[string]interface{}{
			"id":   1,
			"name": "John Doe",
			"email": "john@example.com",
		}

		ec.SetEmployee(ctx, 1, employee)

		var retrieved map[string]interface{}
		exists := ec.GetEmployee(ctx, 1, &retrieved)

		if !exists {
			t.Error("Expected employee to exist")
		}
	})

	t.Run("Invalidate employee", func(t *testing.T) {
		employee := map[string]interface{}{"id": 2, "name": "Jane"}
		ec.SetEmployee(ctx, 2, employee)

		ec.InvalidateEmployee(ctx, 2)

		var retrieved map[string]interface{}
		exists := ec.GetEmployee(ctx, 2, &retrieved)

		if exists {
			t.Error("Expected employee to be invalidated")
		}
	})
}

// TestCacheSize tests cache size limiting and eviction
func TestCacheSizeLimiting(t *testing.T) {
	c := cache.NewCache(3, 1*time.Hour) // Small cache for testing

	t.Run("Size tracking", func(t *testing.T) {
		c.Set("k1", "v1")
		c.Set("k2", "v2")
		c.Set("k3", "v3")

		if c.Size() != 3 {
			t.Errorf("Expected size 3, got %d", c.Size())
		}
	})

	t.Run("Eviction when full", func(t *testing.T) {
		c.Clear()
		c.Set("k1", "v1")
		c.Set("k2", "v2")
		c.Set("k3", "v3")

		// This should trigger eviction
		c.Set("k4", "v4")

		// Cache size should still be max (3)
		if c.Size() > 3 {
			t.Errorf("Expected size <= 3, got %d", c.Size())
		}
	})
}
