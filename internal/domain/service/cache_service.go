package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gin-boilerplate/internal/infrastructure/redis"
)

type CacheService struct {
	redisClient *redis.RedisClient
	prefix     string
	defaultTTL time.Duration
}

func NewCacheService(redisClient *redis.RedisClient) *CacheService {
	return &CacheService{
		redisClient: redisClient,
		prefix:     "gin-boilerplate:",
		defaultTTL:  15 * time.Minute, // 15 minutes default TTL
	}
}

// CacheKey represents a cache key with namespace and identifier
type CacheKey struct {
	Namespace string
	ID        string
}

// String returns formatted cache key
func (ck CacheKey) String() string {
	return fmt.Sprintf("%s:%s:%s", ck.Namespace, ck.ID)
}

// Set stores a value in cache with TTL
func (s *CacheService) Set(ctx context.Context, key CacheKey, value interface{}, ttl ...time.Duration) error {
	cacheKey := key.String()

	// Use provided TTL or default TTL
	expiration := s.defaultTTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	}

	// Serialize value if it's not a string
	var serializedValue interface{}
	if strVal, ok := value.(string); ok {
		serializedValue = strVal
	} else {
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal cache value: %w", err)
		}
		serializedValue = string(jsonBytes)
	}

	return s.redisClient.Set(ctx, cacheKey, serializedValue, expiration)
}

// Get retrieves a value from cache
func (s *CacheService) Get(ctx context.Context, key CacheKey, dest interface{}) error {
	cacheKey := key.String()

	val, err := s.redisClient.Get(ctx, cacheKey)
	if err != nil {
		return err
	}
	if val == "" {
		return nil // Cache miss
	}

	// Try to unmarshal to destination type if it's not a string
	if _, ok := dest.(string); !ok {
		return json.Unmarshal([]byte(val), dest)
	} else {
		// For string destination, assign directly
		strPtr := dest.(*string)
		*strPtr = val
	}

	return nil
}

// GetString retrieves a string value from cache
func (s *CacheService) GetString(ctx context.Context, key CacheKey) (string, error) {
	cacheKey := key.String()

	val, err := s.redisClient.Get(ctx, cacheKey)
	if err != nil {
		return "", err
	}
	if val == "" {
		return "", nil // Cache miss
	}

	return val, nil
}

// Delete removes a value from cache
func (s *CacheService) Delete(ctx context.Context, key CacheKey) error {
	cacheKey := key.String()
	return s.redisClient.Del(ctx, cacheKey)
}

// DeletePattern removes all keys matching a pattern
func (s *CacheService) DeletePattern(ctx context.Context, pattern string) error {
	pattern = fmt.Sprintf("%s*%s", s.prefix, pattern)

	// This is a simplified implementation
	// In production, you might want to use SCAN for large datasets
	keys, err := s.redisClient.GetClient().Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil // No keys to delete
	}

	// Delete all matching keys
	for _, key := range keys {
		if err := s.redisClient.Del(ctx, key).Err(); err != nil {
			// Log error but continue with other keys
			fmt.Printf("Warning: failed to delete cache key %s: %v\n", key, err)
		}
	}

	return nil
}

// SetWithExpiration stores a value with custom expiration
func (s *CacheService) SetWithExpiration(ctx context.Context, key CacheKey, value interface{}, expiration time.Duration) error {
	cacheKey := key.String()

	// Serialize value if it's not a string
	var serializedValue interface{}
	if strVal, ok := value.(string); ok {
		serializedValue = strVal
	} else {
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal cache value: %w", err)
		}
		serializedValue = string(jsonBytes)
	}

	return s.redisClient.Set(ctx, cacheKey, serializedValue, expiration)
}

// Exists checks if a key exists in cache
func (s *CacheService) Exists(ctx context.Context, key CacheKey) (bool, error) {
	cacheKey := key.String()
	return s.redisClient.Exists(ctx, cacheKey)
}

// Increment atomic increment for counters
func (s *CacheService) Increment(ctx context.Context, key CacheKey) (int64, error) {
	cacheKey := key.String()
	return s.redisClient.Increment(ctx, cacheKey)
}

// Utility functions for common cache namespaces
func UserCacheKey(userID string) CacheKey {
	return CacheKey{Namespace: "user", ID: userID}
}

func AuthCacheKey(token string) CacheKey {
	return CacheKey{Namespace: "auth", ID: token}
}

func RateLimitCacheKey(identifier string) CacheKey {
	return CacheKey{Namespace: "rate_limit", ID: identifier}
}

func DocumentCacheKey(documentID string) CacheKey {
	return CacheKey{Namespace: "document", ID: documentID}
}

func SessionCacheKey(sessionID string) CacheKey {
	return CacheKey{Namespace: "session", ID: sessionID}
}