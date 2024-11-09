package mocks

import (
	"errors"
	"time"
)

type MockRedis struct {
	data      map[string]string
	isHealthy bool
}

func NewMockRedis() *MockRedis {
	return &MockRedis{
		data:      make(map[string]string),
		isHealthy: true,
	}
}

func (m *MockRedis) Ping() error {
	if !m.isHealthy {
		return errors.New("redis server is not healthy")
	}
	return nil
}

func (m *MockRedis) Close() error {
	m.isHealthy = false
	return nil
}

func (m *MockRedis) IsHealthy() bool {
	return m.isHealthy
}

func (m *MockRedis) Set(key string, value interface{}, ttl time.Duration) error {
	strValue, ok := value.(string)
	if !ok {
		return errors.New("value must be a string")
	}
	m.data[key] = strValue
	return nil
}

func (m *MockRedis) Get(key string) (string, error) {
	value, exists := m.data[key]
	if !exists {
		return "", errors.New("key not found")
	}

	return value, nil
}

func (m *MockRedis) Del(key string) error {
	if !m.isHealthy {
		return errors.New("redis server is not healthy")
	}
	delete(m.data, key)
	return nil
}
