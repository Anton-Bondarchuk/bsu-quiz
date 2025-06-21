package repository

import (
	"bsu-quiz/internal/config"
	"bsu-quiz/internal/domain/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStorage implements Storage using Redis
type RedisStorage struct {
	client        *redis.Client
	keyPrefix     string
	defaultExpiry time.Duration
}

type RdsOptions struct {
	prefix        string
	defaultExpiry time.Duration
}

type RdsOption func(*RdsOptions)

func WithRdsPrefix(prefix string) RdsOption {
	return func(args *RdsOptions) {
		args.prefix = prefix
	}
}

func WithRdsDefaultExpiry(defaultExpiry time.Duration) RdsOption {
	return func(args *RdsOptions) {
		args.defaultExpiry = defaultExpiry
	}
}

func NewRedisStorage(config config.RedisConfig, setters ...RdsOption) *RedisStorage {
	opt := &RdsOptions{
		prefix: "fsm:",
		defaultExpiry: 24 * time.Hour,
	}

	for _, set := range setters {
		set(opt)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	return &RedisStorage{
		client:        client,
		keyPrefix:     opt.prefix,
		defaultExpiry: opt.defaultExpiry,
	}
}

// makeStateKey creates a key for storing a state
func (s *RedisStorage) makeStateKey(chatID, userID int64) string {
	return fmt.Sprintf("%sstate:%d:%d", s.keyPrefix, chatID, userID)
}

// makeDataKey creates a key for storing data
func (s *RedisStorage) makeDataKey(chatID, userID int64) string {
	return fmt.Sprintf("%sdata:%d:%d", s.keyPrefix, chatID, userID)
}

// Get implements Storage.Get
func (s *RedisStorage) Get(ctx context.Context, chatID int64, userID int64) (models.State, error) {
	key := s.makeStateKey(chatID, userID)
	val, err := s.client.Get(ctx, key).Result()

	if err == redis.Nil {
		// No state found, return empty state without error
		return models.DefaultState, nil
	} else if err != nil {
		return models.DefaultState, fmt.Errorf("failed to get state from Redis: %w", err)
	}

	return models.State(val), nil
}

// Set implements Storage.Set
func (s *RedisStorage) Set(ctx context.Context, chatID int64, userID int64, state models.State) error {
	key := s.makeStateKey(chatID, userID)
	err := s.client.Set(ctx, key, string(state), s.defaultExpiry).Err()
	if err != nil {
		return fmt.Errorf("failed to set state in Redis: %w", err)
	}
	return nil
}

// Delete implements Storage.Delete
func (s *RedisStorage) Delete(ctx context.Context, chatID int64, userID int64) error {
	key := s.makeStateKey(chatID, userID)
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete state from Redis: %w", err)
	}
	return nil
}

// GetData implements Storage.GetData
func (s *RedisStorage) GetData(ctx context.Context, chatID int64, userID int64, key string) (interface{}, error) {
	dataKey := s.makeDataKey(chatID, userID)
	val, err := s.client.HGet(ctx, dataKey, key).Result()

	if err == redis.Nil {
		// No data found for this key
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get data from Redis: %w", err)
	}

	// Unmarshal the JSON data
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data from Redis: %w", err)
	}

	return result, nil
}

// SetData implements Storage.SetData
func (s *RedisStorage) SetData(ctx context.Context, chatID int64, userID int64, key string, value interface{}) error {
	dataKey := s.makeDataKey(chatID, userID)

	// Marshal the value to JSON
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal data for Redis: %w", err)
	}

	// Set the data in Redis hash
	if err := s.client.HSet(ctx, dataKey, key, jsonData).Err(); err != nil {
		return fmt.Errorf("failed to set data in Redis: %w", err)
	}

	// Set expiry for the hash
	if err := s.client.Expire(ctx, dataKey, s.defaultExpiry).Err(); err != nil {
		return fmt.Errorf("failed to set expiry for Redis key: %w", err)
	}

	return nil
}

// ClearData implements Storage.ClearData
func (s *RedisStorage) ClearData(ctx context.Context, chatID int64, userID int64) error {
	dataKey := s.makeDataKey(chatID, userID)
	err := s.client.Del(ctx, dataKey).Err()
	if err != nil {
		return fmt.Errorf("failed to clear data from Redis: %w", err)
	}
	return nil
}

// Close closes the Redis client connection
func (s *RedisStorage) Close() error {
	return s.client.Close()
}

// Ping tests the connection to Redis
func (s *RedisStorage) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}
