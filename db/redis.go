package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

type RedisInterface interface {
	Ping() error
	Close() error
	IsHealthy() bool
	Set(key string, value interface{}, ttl time.Duration) error
	Get(key string) (string, error)
	Del(key string) error
}

var Redis RedisClient

type RedisClient struct {
	client    *redis.Client
	isHealthy bool
}

func InitRedis() {
	err := connectRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	go startRedisHealthCheck()
}

func connectRedis() error {
	addr := os.Getenv("REDIS_ADDR")
	password := os.Getenv("REDIS_PASSWORD")

	Redis.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if err := Redis.client.Ping().Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis!!")
	Redis.isHealthy = true
	return nil
}

func startRedisHealthCheck() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if err := Redis.client.Ping().Err(); err != nil {
			Redis.isHealthy = false
			log.Printf("Redis ping failed: %v", err)

			if err := connectRedis(); err != nil {
				log.Printf("Failed to reconnect to Redis: %v", err)
			} else {
				Redis.isHealthy = true
				log.Println("Reconnected to Redis!")
			}
		}
	}
}

func (r *RedisClient) IsHealthy() bool {
	return r.isHealthy
}

func (r *RedisClient) Ping() error {
	return r.client.Ping().Err()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Set(key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(key, value, ttl).Err()
}

func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(key).Result()
}

func (r *RedisClient) Del(key string) error {
	return r.client.Del(key).Err()
}
