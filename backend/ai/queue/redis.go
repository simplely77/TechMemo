package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const redisQueueKey = "techmemo:ai:tasks"

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(addr, password string, db int) *RedisQueue {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisQueue{client: client}
}

func (r *RedisQueue) Publish(task AITask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("序列化任务失败: %w", err)
	}
	return r.client.LPush(context.Background(), redisQueueKey, data).Err()
}

func (r *RedisQueue) Consume() (<-chan AITask, error) {
	ch := make(chan AITask)
	go func() {
		for {
			// BRPOP 阻塞等待，0 表示永不超时
			result, err := r.client.BRPop(context.Background(), 0, redisQueueKey).Result()
			if err != nil {
				continue
			}
			var task AITask
			if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
				continue
			}
			ch <- task
		}
	}()
	return ch, nil
}
